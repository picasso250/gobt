package gobt

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
)

const doNotBotherTracker = true // for debug use

// const
const hashSize = 20
const peerIDSize = 20

// settings
var maxPeerCount int

// DownloadRoot root directory of download
var DownloadRoot string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	maxPeerCount = *flag.Int("max-peer-count", 30, "how many peers to connect")
	DownloadRoot = *flag.String("root", ".", "download root directory")

	myPeerID = genPeerID()
	gPeersToStart = make(chan *peer, 10)

}

var byteTable = map[byte]int{
	0:   0,
	1:   1,
	2:   1,
	3:   2,
	4:   1,
	5:   2,
	6:   2,
	7:   3,
	8:   1,
	9:   2,
	10:  2,
	11:  3,
	12:  2,
	13:  3,
	14:  3,
	15:  4,
	16:  1,
	17:  2,
	18:  2,
	19:  3,
	20:  2,
	21:  3,
	22:  3,
	23:  4,
	24:  2,
	25:  3,
	26:  3,
	27:  4,
	28:  3,
	29:  4,
	30:  4,
	31:  5,
	32:  1,
	33:  2,
	34:  2,
	35:  3,
	36:  2,
	37:  3,
	38:  3,
	39:  4,
	40:  2,
	41:  3,
	42:  3,
	43:  4,
	44:  3,
	45:  4,
	46:  4,
	47:  5,
	48:  2,
	49:  3,
	50:  3,
	51:  4,
	52:  3,
	53:  4,
	54:  4,
	55:  5,
	56:  3,
	57:  4,
	58:  4,
	59:  5,
	60:  4,
	61:  5,
	62:  5,
	63:  6,
	64:  0,
	65:  1,
	66:  1,
	67:  2,
	68:  1,
	69:  2,
	70:  2,
	71:  3,
	72:  1,
	73:  2,
	74:  2,
	75:  3,
	76:  2,
	77:  3,
	78:  3,
	79:  4,
	80:  1,
	81:  2,
	82:  2,
	83:  3,
	84:  2,
	85:  3,
	86:  3,
	87:  4,
	88:  2,
	89:  3,
	90:  3,
	91:  4,
	92:  3,
	93:  4,
	94:  4,
	95:  5,
	96:  1,
	97:  2,
	98:  2,
	99:  3,
	100: 2,
	101: 3,
	102: 3,
	103: 4,
	104: 2,
	105: 3,
	106: 3,
	107: 4,
	108: 3,
	109: 4,
	110: 4,
	111: 5,
	112: 2,
	113: 3,
	114: 3,
	115: 4,
	116: 3,
	117: 4,
	118: 4,
	119: 5,
	120: 3,
	121: 4,
	122: 4,
	123: 5,
	124: 4,
	125: 5,
	126: 5,
	127: 6,
	128: 0,
	129: 1,
	130: 1,
	131: 2,
	132: 1,
	133: 2,
	134: 2,
	135: 3,
	136: 1,
	137: 2,
	138: 2,
	139: 3,
	140: 2,
	141: 3,
	142: 3,
	143: 4,
	144: 1,
	145: 2,
	146: 2,
	147: 3,
	148: 2,
	149: 3,
	150: 3,
	151: 4,
	152: 2,
	153: 3,
	154: 3,
	155: 4,
	156: 3,
	157: 4,
	158: 4,
	159: 5,
	160: 1,
	161: 2,
	162: 2,
	163: 3,
	164: 2,
	165: 3,
	166: 3,
	167: 4,
	168: 2,
	169: 3,
	170: 3,
	171: 4,
	172: 3,
	173: 4,
	174: 4,
	175: 5,
	176: 2,
	177: 3,
	178: 3,
	179: 4,
	180: 3,
	181: 4,
	182: 4,
	183: 5,
	184: 3,
	185: 4,
	186: 4,
	187: 5,
	188: 4,
	189: 5,
	190: 5,
	191: 6,
	192: 0,
	193: 1,
	194: 1,
	195: 2,
	196: 1,
	197: 2,
	198: 2,
	199: 3,
	200: 1,
	201: 2,
	202: 2,
	203: 3,
	204: 2,
	205: 3,
	206: 3,
	207: 4,
	208: 1,
	209: 2,
	210: 2,
	211: 3,
	212: 2,
	213: 3,
	214: 3,
	215: 4,
	216: 2,
	217: 3,
	218: 3,
	219: 4,
	220: 3,
	221: 4,
	222: 4,
	223: 5,
	224: 1,
	225: 2,
	226: 2,
	227: 3,
	228: 2,
	229: 3,
	230: 3,
	231: 4,
	232: 2,
	233: 3,
	234: 3,
	235: 4,
	236: 3,
	237: 4,
	238: 4,
	239: 5,
	240: 2,
	241: 3,
	242: 3,
	243: 4,
	244: 3,
	245: 4,
	246: 4,
	247: 5,
	248: 3,
	249: 4,
	250: 4,
	251: 5,
	252: 4,
	253: 5,
	254: 5,
	255: 6,
}

const (
	statePeerInit = iota
	statePeerConnected
	statePeerError
)

var myPeerID peerID
var gBitField *bitfield
var peersStateMap map[string]int
var peersMap map[string]*peer
var peersMapMutex sync.RWMutex
var gPeersToStart chan *peer

type ipPort struct {
	IP   uint32
	Port uint16
}

func newIPPortFromUint64(u uint64) ipPort {
	return ipPort{
		uint32(0xFFFFFFFF & u),
		uint16(u >> 32),
	}
}

func (i ipPort) Uint64() uint64 {
	return uint64(i.IP)<<32 | uint64(i.Port)
}

func (i ipPort) String() string {
	return IPIntToString(int(i.IP)) + ":" + strconv.Itoa(int(i.Port))
}

// Download download BT file
func Download(filename string) {

	metaInfo, err := NewMetainfoFromFile(filename)
	if err != nil {
		fmt.Printf("parse bt file error: %s\n", err)
		return
	}

	gBitField, err = ensureFile(metaInfo.Info)
	if err != nil {
		fmt.Printf("file error: %s\n", err)
		log.Fatal(err)
	}

	ln, port, err := availablePort()
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	trackerProtocol(metaInfo, port)

	go func() {
		fmt.Printf("Listening...\n")
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("connection comes from %s\n", conn.RemoteAddr())
			go handleConnection(conn, metaInfo)
		}
	}()

	// do peers
	for {
		peer := <-gPeersToStart

		peersMapMutex.RLock()
		if peersMap[peer.String()] == nil {
			fmt.Printf("start peer %s\n", peer.String())
			peersMap[peer.String()] = peer
			go peer.startHandle(metaInfo)
		}
		peersMapMutex.RUnlock()
	}

}
func handleConnection(conn net.Conn, metaInfo *Metainfo) {
	addr := conn.RemoteAddr()
	peersMapMutex.RLock()
	defer peersMapMutex.RUnlock()
	if peersMap[addr.String()] == nil {
		var pid peerID
		p := newPeer(addr, pid)
		p.Conn = conn
		go p.handleConnection(metaInfo)
	}
}

func compactPeerList(b []byte, piecesCount int) ([]*peer, error) {
	ret := make([]*peer, 0)
	if len(b)%6 != 0 {
		return nil, errors.New("compact peers is not 6s")
	}
	for i := 0; i < len(b)/6; i++ {
		ip := int(b[i*6])*0xFFFFFF + int(b[i*6+1])*0xFFFF + int(b[i*6+2])*0xFF + int(b[i*6+3])
		port := int(b[i*6+4])*0xFF + int(b[i*6+5])

		address := IPIntToString(ip) + ":" + strconv.Itoa(port)
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			fmt.Printf("peer address resolve error: %s\n", err)
			continue
		}

		var pid peerID
		i := newPeer(addr, pid)
		ret = append(ret, i)
	}
	return ret, nil
}

func peerList(peers map[string]interface{}, piecesCount int) ([]*peer, error) {
	ret := make([]*peer, 0)
	for _, p := range peers {
		pm := p.(map[string]interface{})

		pid, err := peerIDFromBytes(pm["peer id"].([]byte))
		if err != nil {
			return ret, err
		}

		address := string(pm["ip"].([]byte)) + ":" + string(pm["port"].([]byte))
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			fmt.Printf("peer address resolve error: %s\n", err)
			continue
		}
		pp := newPeer(addr, pid)
		ret = append(ret, pp)
	}
	return ret, nil
}

func getAllAnnounce(metainfo *Metainfo) (ret []string) {
	return unique(append(metainfo.AnnounceList, metainfo.Announce))
}
func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func udpTracker(address string, metainfo *Metainfo) error {
	// http://www.bittorrent.org/beps/bep_0015.html
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	var transactionID = uint32(rand.Int31())
	err = connectRequest(conn, transactionID)
	if err != nil {
		return err
	}

	connectionID, err := connectResponse(conn, transactionID)
	if err != nil {
		return err
	}

	// IPv4 announce request
	req := NewTrackerRequest(metainfo, 88) // TODO what port?
	announceRequest(conn, transactionID, connectionID, req)

	// IPv4 announce response
	announceResponse(conn, transactionID)
	return nil
}
func announceResponse(conn *net.UDPConn, transactionID uint32) (*TrackerResponse, error) {

	action, err := readUint32(conn)
	if err != nil {
		return nil, err
	}
	// action          1 // announce
	if action != 1 {
		log.Fatal("action not 1")
	}

	t, err := readUint32(conn)
	if err != nil {
		return nil, err
	}
	if t != transactionID {
		log.Fatal("transactionID mismatch")
	}

	interval, err := readUint32(conn)
	if err != nil {
		return nil, err
	}
	fmt.Printf("interval %d\n", interval)

	leechers, err := readUint32(conn)
	if err != nil {
		return nil, err
	}
	fmt.Printf("leechers %d\n", leechers)

	seeders, err := readUint32(conn)
	if err != nil {
		return nil, err
	}
	fmt.Printf("seeders %d\n", seeders)

	lst := make([]ipPort, 0)
	for {
		ip, err := readUint32(conn)
		if err != nil {
			return nil, err
		}
		port, err := read2Byte(conn)
		if err != nil {
			if err == io.EOF {
				lst = append(lst, ipPort{ip, port})
				break
			}
			return nil, err
		}
		lst = append(lst, ipPort{ip, port})
	}
	resp := TrackerResponse{
		Action:        action,
		TransactionID: t,
		Interval:      interval,
		Leechers:      leechers,
		Seeders:       seeders,
		IPPort:        lst,
	}
	return &resp, nil
}
func announceRequest(conn *net.UDPConn, transactionID uint32, connectionID uint64, req *TrackerRequest) error {
	b, err := announceRequestBytes(conn, transactionID, connectionID, req)
	if err != nil {
		return err
	}
	n, err := conn.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		log.Fatal("only n", n)
	}
	return nil
}

func announceRequestBytes(conn *net.UDPConn, transactionID uint32, connectionID uint64, req *TrackerRequest) ([]byte, error) {
	action := uint32(1) // announce
	var err error
	buf := bytes.NewBuffer(make([]byte, 1024))

	err = binary.Write(buf, binary.BigEndian, connectionID)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, action)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, transactionID)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.InfoHash)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.PeerID)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.Downloaded)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.Left)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.Uploaded)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.Event)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.IP)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.Key)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.NumWant)
	if err != nil {
		return nil, err
	}

	binary.Write(buf, binary.BigEndian, req.Port)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func genPeerID() [peerIDSize]byte {
	b := make([]byte, peerIDSize)
	n, err := rand.Read(b)
	if err != nil {
		log.Fatal("rand error")
	}
	if n != peerIDSize {
		log.Fatal("rand length error")
	}
	var a [peerIDSize]byte
	copy(a[:], b[:peerIDSize])
	return a
}
func infoHash(info map[string]interface{}) hash {
	s, err := Encode(info)
	if err != nil {
		log.Fatal("info encode fail")
	}
	return sha1.Sum(s)
}
func connectResponse(conn *net.UDPConn, transactionID uint32) (uint64, error) {

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return 0, err
	}
	fmt.Printf("read %d bytes\n", n)
	if n != 16 {
		log.Fatal("not 16 read")
	}
	buf := bytes.NewReader(buffer)
	action := uint32(0)
	err = binary.Read(buf, binary.BigEndian, &action)
	if err != nil {
		return 0, err
	}
	if action != 0 {
		log.Fatal("action not 0")
	}
	transID := uint32(0)
	err = binary.Read(buf, binary.BigEndian, &transID)
	if err != nil {
		return 0, err
	}
	if transID != transactionID {
		return 0, errors.New("transaction_id not equal")
	}
	var connectionID uint64
	err = binary.Read(buf, binary.BigEndian, &connectionID)
	if err != nil {
		return 0, err
	}
	return connectionID, nil
}
func connectRequest(conn *net.UDPConn, transactionID uint32) error {
	b, err := connectRequestBytes(transactionID)
	if err != nil {
		return err
	}
	n, err := conn.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		log.Fatal("only n", n)
	}
	return nil
}
func connectRequestBytes(transactionID uint32) ([]byte, error) {
	var protocolID uint64 = 0x41727101980 // magic constant
	action := uint32(0)
	var err error
	buf := bytes.NewBuffer(make([]byte, 16))
	err = binary.Write(buf, binary.BigEndian, protocolID)
	if err != nil {
		return nil, err
	}
	binary.Write(buf, binary.BigEndian, action)
	if err != nil {
		return nil, err
	}
	binary.Write(buf, binary.BigEndian, transactionID)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
