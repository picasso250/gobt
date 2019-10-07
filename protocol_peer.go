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
	"strconv"
	"time"
)

const requestLength = uint32(1 << 14) // All current implementations use 2^14 (16 kiB)

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
	IP     uint32
	Port   uint16
	PeerID peerID
	// state
	AmChoking      uint32 // 本客户端正在choke远程peer。
	AmInterested   uint32 // 本客户端对远程peer感兴趣。
	PeerChoking    uint32 // 远程peer正choke本客户端。
	PeerInterested uint32 // 远程peer对本客户端感兴趣。

	Conn        net.Conn
	pieceOffset int32
	Bitfield    bitfield
}

func newPeer(ip uint32, port uint16, pid peerID, bitfieldSize int) *peer {
	return &peer{
		IP:     ip,
		Port:   port,
		PeerID: pid,
		// 客户端连接开始时状态是choke和not interested(不感兴趣)。换句话就是：
		AmChoking:      1,
		AmInterested:   0,
		PeerChoking:    1,
		PeerInterested: 0,

		Conn:     nil,
		Bitfield: allZeroBitField(bitfieldSize),
	}
}
func (p *peer) Uint64() uint64 {
	return uint64(p.IP)<<32 | uint64(p.Port)
}

func (p *peer) String() string {
	return IPIntToString(int(p.IP)) + ":" + strconv.Itoa(int(p.Port))
}

func (p *peer) start(metainfo *Metainfo) {
	var err error

	// maybe we don't need to lock here, but who knows
	peersMapMutex.Lock()
	p.Conn, err = net.Dial("tcp4", p.String())
	peersMapMutex.Unlock()
	if err != nil {
		fmt.Printf("dial tcp %s error: %s\n", p.String(), err)
		return
	}
	defer p.Conn.Close()

	err = handshake(p, metainfo)
	if err != nil {
		fmt.Printf("%s handshake error: %s", p, err)
		return
	}

	heartBeatChan := make(chan int)
	go heartBeat(p.Conn, heartBeatChan)
	defer func() {
		// inform heart beat to stop
		heartBeatChan <- 1
	}()

	err = p.peerMessages(metainfo.Info)
	if err != nil {
		fmt.Printf("peer messages error: %s\n", err)
		return
	}
}

func (p *peer) peerMessages(info *MetainfoInfo) error {
	conn := p.Conn

	msg, err := buildpeerMessageBitfield(info)
	if err != nil {
		return err
	}
	err = sendMessage(conn, msg)
	if err != nil {
		return err
	}

	// for simplicity we are interested in every one and do not choke anyone
	err = sendCmd(conn, typeUnchoke)
	if err != nil {
		return err
	}
	p.AmChoking = 0

	err = sendCmd(conn, typeInterested)
	if err != nil {
		return err
	}
	p.AmInterested = 1

	return p.loop()
}
func (p *peer) loop() error {
	for {
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
			if err != nil {
				return err
			}
		case typeRequest:
			err = p.doRequest(b, info)
			if err != nil {
				return err
			}
		}
	}
}
func (p *peer) doRequest(b []byte, info *MetainfoInfo) error {
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
		b, err := readSome(info, int(index), int64(begin), int64(length))
		if err != nil {
			return err
		}

	}

	//todo move to piece
	piece := buf.Bytes()
	err = writeToFile(info, int(index), int64(begin), piece)
	if err != nil {
		return err
	}
	return nil
}
func (p *peer) doBitfield(b []byte) error {
	if len(b) != len(gBitField) {
		return errors.New("bitfield length mismatch")
	}
	p.Bitfield = b
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
func sendCmd(conn net.Conn, t uint32) error {
	err := writeInteger(conn, uint32(4))
	if err != nil {
		return err
	}
	return writeInteger(conn, t)
}
func alternatingStream(conn net.Conn, info *MetainfoInfo) error {
	// randomly pick a pieace to download
	index := randIndex(info)
	sendMessage(conn, requestMessage(int32(index)))
	return nil
}

// return -1 if all bit set
func randIndex(info *MetainfoInfo) int {
	cnt := info.piecesCount()
	start := rand.Intn(cnt)
	for i := 0; i < cnt; i++ {
		ii := (start + i) % cnt
		gBitFieldMutex.RLock()
		if gBitField.Bit(ii) == 0 {
			return ii
		}
		gBitFieldMutex.RUnlock()
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
func buildpeerMessageBitfield(info *MetainfoInfo) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := writeInteger(buf, uint32(typeBitfield))
	if err != nil {
		return nil, err
	}
	b, err := info.bitfield()
	if err != nil {
		return nil, err
	}
	n, err := buf.Write(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		log.Fatal("write not enough bytes")
	}
	return buf, nil
}

func handshake(p *peer, metainfo *Metainfo) error {
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

func heartBeat(conn net.Conn, ch chan int) {
	for {
		select {
		case <-time.After(2 * time.Minute):
			err := writeInteger(conn, 0)
			if err != nil {
				fmt.Printf("heart beat error: %s", err)
				return
			}
		case <-ch:
			return
		}
	}
}
