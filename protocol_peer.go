package gobt

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

const requestLength = uint32(1 << 14) // All current implementations use 2^14 (16 kiB)
const chanWaitTimeout = time.Millisecond * 10

// non-keepalive messages start with a single byte which gives their type
const (
	typeChoke uint32 = iota
	typeUnchoke
	typeInterested
	typeNotInterested
	typeHave
	typeBitfield
	typeRequest
	typePiece
	typeCancel
)

type peer struct {
	Addr   net.Addr
	PeerID peerID
	// state
	AmChoking      uint32 // 本客户端正在choke远程peer。
	AmInterested   uint32 // 本客户端对远程peer感兴趣。
	PeerChoking    uint32 // 远程peer正choke本客户端。
	PeerInterested uint32 // 远程peer对本客户端感兴趣。

	Conn           net.Conn
	Bitfield       *bitfield
	Cancel         chan iblPack
	WillCancel     []iblPack
	PieceOffsetMap map[uint32]int     // piece start 0----piece offset----piece end
	ToSend         chan messageToSend // send to peer
	Error          chan error
}
type messageToSend struct {
	Request iblPack
	Message []byte
}

type iblPack []byte // pack index, begin, and length to bytes

func newPeer(addr net.Addr, pid peerID) *peer {
	return &peer{
		Addr:   addr,
		PeerID: pid,
		// 客户端连接开始时状态是choke和not interested(不感兴趣)。换句话就是：
		AmChoking:      1,
		AmInterested:   0,
		PeerChoking:    1,
		PeerInterested: 0,

		Conn:     nil, // Multiple goroutines may invoke methods on a Conn simultaneously
		Bitfield: allZeroBitField(gBitField.Len()),
		Cancel:   make(chan iblPack, 10),
		ToSend:   make(chan messageToSend), // for simplicity, make it sync
	}
}

func (p *peer) String() string {
	return p.Addr.String()
}

func (p *peer) handleConnection(metainfo *Metainfo) {
	defer p.Conn.Close()
	p.startListen(metainfo)
}
func (p *peer) startHandle(metainfo *Metainfo) {
	var err error

	// maybe we don't need to lock here, but who knows
	peersMapMutex.Lock()
	// todo tcp v6
	p.Conn, err = net.Dial("tcp4", p.Addr.String())
	peersMapMutex.Unlock()
	if err != nil {
		fmt.Printf("dial tcp %s error: %s\n", p.Addr.String(), err)
		return
	}

	p.handleConnection(metainfo)

}
func (p *peer) startListen(metainfo *Metainfo) {
	var err error

	err = handshake(p, metainfo)
	if err != nil {
		fmt.Printf("%s handshake error: %s", p, err)
		return
	}

	heartBeatWillStop := make(chan int)
	go p.heartBeat(heartBeatWillStop)
	defer func() {
		// inform heart beat to stop
		heartBeatWillStop <- 1
	}()

	err = p.peerMessages(metainfo.Info)
	if err != nil {
		fmt.Printf("peer messages error: %s\n", err)
		return
	}
}

func (p *peer) peerMessages(info *MetainfoInfo) error {

	// start to send
	go p.startSend()

	msg, err := buildPeerMessageBitfield(info)
	if err != nil {
		return err
	}
	p.ToSend <- messageToSend{nil, msg}

	// for simplicity we are interested in every one and do not choke anyone
	p.sendCmd(typeUnchoke)
	p.AmChoking = 0

	p.sendCmd(typeInterested)
	p.AmInterested = 1

	return p.loop(info)
}

func (p *peer) startSend() {
	// these two are goroutine safe
	conn := p.Conn

	for {

		select {

		case msg := <-p.ToSend:
			if is, willCancel := inCancel(msg.Request, p.WillCancel); is {
				// drop this message
				p.WillCancel = willCancel
			} else {
				err := writeMessage(conn, msg.Message)
				if err != nil {
					p.Error <- err
					return
				}
			}

		case c := <-p.Cancel:
			p.WillCancel = append(p.WillCancel, c)
		}

	}
}

func writeMessage(conn net.Conn, msg []byte) error {
	b := msgWrap(msg)
	return writeAll(conn, b)
}
func inCancel(elem iblPack, list []iblPack) (bool, []iblPack) {
	for i, v := range list {
		if bytes.Equal(v, elem) {
			return true, append(list[:i], list[i+1:]...)
		}
	}
	return false, list
}

func packCancel(index int32, begin int32, length int32) []byte {
	return packMessage(nil, index, begin, length)
}
func msgWrap(msg []byte) []byte {
	buf := new(bytes.Buffer)
	err := writeAll(buf, msg)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
func (p *peer) loop(info *MetainfoInfo) (err error) {
	errOccur := make(chan error)
	for {
		select {
		case err := <-errOccur:
			return err
		case <-time.After(chanWaitTimeout):
			// go on
		}

		if p.PeerChoking == 0 {
			// send him message for request
			// it is safe to goroutine
			go func() {
				// todo tit-for-tat-ish algorithm
				err = requestPeer(p, info)
				if err != nil {
					errOccur <- err
				}
			}()
		}

		if p.AmInterested == 0 {
			time.Sleep(time.Second)
			continue
		}
		t, b, err := readNextMsg(p.Conn)
		if err != nil {
			return err
		}
		switch t {
		default:
			return errors.New("unknown message type")
		case typeChoke:
			p.PeerChoking = 1
		case typeUnchoke:
			p.PeerChoking = 0
		case typeInterested:
			p.PeerInterested = 1
		case typeNotInterested:
			p.PeerInterested = 0
		case typeHave:
			err = p.doHave(b)
			if err != nil {
				return err
			}
		case typeBitfield:
			err = p.doBitfield(b)
		case typeRequest:
			err = p.doRequest(b, info, errOccur)
		case typeCancel:
			p.doCancel(b)
		case typePiece:
			err = p.doPiece(b, info)
		}
		if err != nil {
			return err
		}
	}
}

func (p *peer) doPiece(b []byte, info *MetainfoInfo) (err error) {

	buf := bytes.NewBuffer(b)

	index, err := readUint32(buf)
	if err != nil {
		return err
	}

	begin, err := readUint32(buf)
	if err != nil {
		return err
	}

	piece := buf.Bytes()

	if gBitField.Bit(int(index)) == 1 {
		fmt.Printf("duplicate piece\n")
		return nil
	}

	err = writeToFile(info, int(index), int64(begin), piece)
	if err != nil {
		return err
	}

	p.PieceOffsetMap[index] = int(begin + uint32(len(piece)))
	if int(begin)+len(piece) == info.PieceLength {
		// 校验
		var h hash
		n := copy(h[:], info.Pieces[index*hashSize:(index+1)*hashSize])
		if n != hashSize {
			log.Fatal("copy hash size error")
		}
		isValid, err := checkHash(info, int(index), h)
		if err != nil {
			return err
		}
		if isValid {
			gBitField.SetBit(int(index), 1)
			err = gBitField.ToFile(info.infoFilename())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *peer) doCancel(b []byte) {
	p.Cancel <- b
}
func (p *peer) doRequest(b []byte, info *MetainfoInfo, errOccur chan error) error {
	cancelFlag := b[:3*4]

	buf := bytes.NewBuffer(b)
	index, err := readUint32(buf)
	if err != nil {
		return err
	}
	begin, err := readUint32(buf)
	if err != nil {
		return err
	}
	length, err := readUint32(buf)
	if err != nil {
		return err
	}

	// if we have
	if gBitField.Bit(int(index)) == 1 {

		piece, err := readSomeFileContent(info, int(index), int64(begin), int64(length))
		if err != nil {
			return err
		}
		// 'piece' messages contain an index, begin, and piece
		b := packMessage(piece, int32(typePiece), int32(begin), int32(length))

		p.ToSend <- messageToSend{cancelFlag, b}

		//todo delete this
		go func() {
			err = p.sendTypeMessageWhile(typePiece, (b))
			if err != nil {
				errOccur <- err
			}
		}()

	}

	return nil
}

func (p *peer) sendTypeMessageWhile(t uint32, msg []byte) error {

	// these two are goroutine safe
	conn := p.Conn
	willCancel := p.Cancel

	err := writeIntegers(conn, uint32(len(msg)+4), t)
	if err != nil {
		return err
	}

	for {

		select {
		case <-willCancel:
			return nil
		case <-time.After(chanWaitTimeout):
			// we go on
		}

		err = conn.SetDeadline(time.Now().Add(time.Millisecond * 100))
		if err != nil {
			return err
		}

		n, err := conn.Write(msg)
		if err != nil {
			if oe, ok := err.(net.Error); ok {
				if !oe.Timeout() {
					return err
				}
			} else {
				return err
			}
		}

		msg = msg[n:]
		if len(msg) == 0 {
			break
		}
	}

	return nil
}

func (p *peer) doBitfield(b []byte) error {
	if len(b) != (gBitField.Len()) {
		return errors.New("bitfield length mismatch")
	}
	p.Bitfield.SetBitData(b)
	return nil
}
func (p *peer) doHave(b []byte) error {
	buf := bytes.NewBuffer(b)
	index, err := readUint32(buf)
	if err != nil {
		return err
	}
	p.Bitfield.SetBit(int(index), 1)
	return nil
}
func readNextMsg(conn net.Conn) (uint32, []byte, error) {
	// Messages of length zero are keepalives, and ignored
	size := 0
	for {
		size, err := readUint32(conn)
		if err != nil {
			return 0, nil, err
		}
		if size != 0 {
			break
		}
	}

	t, err := readUint32(conn)
	if err != nil {
		return 0, nil, err
	}

	b := make([]byte, size-4)
	if size-4 > 0 {
		_, err = io.ReadFull(conn, b)
		if err != nil {
			return 0, nil, err
		}
	}
	return t, b, nil
}
func (p *peer) sendCmd(t uint32) {
	msg := packMessage(nil, int32(t))
	p.ToSend <- messageToSend{nil, msg}
}

func requestPeer(p *peer, info *MetainfoInfo) error {
	// randomly pick a pieace to download
	index := randIndex(info)
	sendMessage(p.Conn, requestMessage(int32(index)))
	return nil
}

// return -1 if all bit set
func randIndex(info *MetainfoInfo) int {
	cnt := info.piecesCount()
	start := rand.Intn(cnt)
	for i := 0; i < cnt; i++ {
		ii := (start + i) % cnt
		if gBitField.Bit(ii) == 0 {
			return ii
		}
	}
	return -1
}

// var pieceInnerIndex int32 = 0

func requestMessage(index int32) *bytes.Buffer {
	b := new(bytes.Buffer)

	err := writeInteger(b, index)
	if err != nil {
		log.Fatal(err)
	}
	pieceInnerIndex := 0
	err = writeInteger(b, pieceInnerIndex)
	if err != nil {
		log.Fatal(err)
	}

	err = writeInteger(b, requestLength)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
func sendMessage(conn net.Conn, msg *bytes.Buffer) error {
	err := writeInteger(conn, uint32(msg.Len()))
	if err != nil {
		return err
	}
	n, err := msg.WriteTo(conn)
	if err != nil {
		return err
	}
	if int(n) != msg.Len() {
		log.Fatal("write message length not enough")
	}
	return nil
}

// pack ints and bytes
func packMessage(b []byte, is ...int32) []byte {
	buf := bytes.NewBuffer(b)
	err := writeIntegers(buf, is)
	if err != nil {
		log.Fatal(err)
	}
	err = writeAll(buf, b)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func sendTypeMessage(conn net.Conn, t uint32, msg *bytes.Buffer) error {
	err := writeIntegers(conn, uint32(msg.Len()+4), t)
	if err != nil {
		return err
	}
	n, err := msg.WriteTo(conn)
	if err != nil {
		return err
	}
	if int(n) != msg.Len() {
		log.Fatal("write message length not enough")
	}
	return nil
}
func buildPeerMessageBitfield(info *MetainfoInfo) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := writeInteger(buf, uint32(typeBitfield))
	if err != nil {
		return nil, err
	}
	b, err := info.bitfield()
	if err != nil {
		return nil, err
	}
	n, err := buf.Write(b.BitData())
	if err != nil {
		return nil, err
	}
	if n != b.Len() {
		log.Fatal("write not enough bytes")
	}
	return buf.Bytes(), nil
}

func handshake(p *peer, metainfo *Metainfo) error {
	fmt.Printf("handshake with peer %s\n", p.Addr.String())
	conn := p.Conn

	// The handshake starts with character ninteen (decimal) followed by the string 'BitTorrent protocol'
	err := protocol(conn)
	if err != nil {
		fmt.Printf("%s protocol error: %s", p, err)
		return err
	}

	err = reservedBytes(conn)
	if err != nil {
		fmt.Printf("%s reserved bytes error: %s", p, err)
		return err
	}

	err = exchangeSha1Hash(conn, metainfo.InfoHash[:])
	if err != nil {
		fmt.Printf("%s exchange hash error: %s", p, err)
		return err
	}

	err = exchangePeerID(conn, myPeerID[:])
	if err != nil {
		fmt.Printf("%s exchange peer id error: %s", p, err)
		return err
	}

	return nil
}

func exchangePeerID(conn net.Conn, peerID []byte) error {

	n, err := conn.Write(peerID)
	if err != nil {
		return err
	}
	if n != len(peerID) {
		log.Fatal("write reserved bytes length error")
	}

	br := make([]byte, peerIDSize)
	_, err = io.ReadFull(conn, br)
	return err
}
func protocol(conn net.Conn) error {
	s := "BitTorrent protocol"
	n, err := fmt.Fprintf(conn, "%c%s", 19, s)
	if err != nil {
		return err
	}
	if n != len(s)+1 {
		log.Fatal("handshake write failed")
	}
	rd := bufio.NewReader(conn)
	len, err := rd.ReadByte()
	if err != nil {
		return err
	}
	if len != 19 {
		log.Fatal("unknown handshake version")
	}
	b := make([]byte, 19)
	_, err = io.ReadFull(rd, b)
	if err != nil {
		return err
	}
	if string(b) != s {
		log.Fatalf("unknown version: %s", string(b))
	}
	return nil
}

func exchangeSha1Hash(conn net.Conn, infoHash []byte) error {
	n, err := conn.Write(infoHash)
	if err != nil {
		return err
	}
	if n != len(infoHash) {
		log.Fatal("write reserved bytes length error")
	}

	br := make([]byte, hashSize)
	_, err = io.ReadFull(conn, br)
	if err != nil {
		return err
	}
	if bytes.Compare(br, infoHash) != 0 {
		log.Fatal("info hash not equal")
	}

	return nil
}
func reservedBytes(conn net.Conn) error {
	b := make([]byte, 8)
	n, err := conn.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		log.Fatal("write reserved bytes length error")
	}

	br := make([]byte, 8)
	_, err = io.ReadFull(conn, br)
	if err != nil {
		return err
	}
	if bytes.Compare(br, b) != 0 {
		log.Fatal("reserved bytes not 0")
	}

	return nil
}

func (p *peer) heartBeat(willStop chan int) {
	for {
		select {
		case <-time.After(2 * time.Minute):
			p.ToSend <- messageToSend{nil, nil}

		case <-willStop:
			return
		}
	}
}
