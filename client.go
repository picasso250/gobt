package gobt

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const infoHashSize = 20
const peerIDSize = 20

var downloadRoot = "d:\\DOWNLOAD\\test"

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

// TrackerRequest Tracker GET requests
type TrackerRequest struct {
	InfoHash   [infoHashSize]byte
	PeerID     [peerIDSize]byte
	IP         uint32
	Port       uint16
	Uploaded   uint64
	Downloaded uint64
	Left       uint64
	Event      uint32
	Key        uint32
	NumWant    int32
}

// TrackerResponse IPv4 announce response
type TrackerResponse struct {
	Action        uint32
	TransactionID uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32
	IPPort        []ipPort
}

type ipPort struct {
	IP   uint32
	Port uint16
}

// Download download BT file
func Download(filename string) {

	peers := make([]ipPort, 0)

	metainfo, err := parseBTFile(filename)
	if err != nil {
		fmt.Printf("parse bt file error: %s\n", err)
		return
	}

	info := metainfo["info"].(map[string]interface{})
	err = ensureFile(info)
	if err != nil {
		fmt.Printf("file error: %s\n", err)
	}

	if metainfo["announce"] == nil {
		fmt.Printf("no announce key\n")
		return
	}
	allAnnounce := getAllAnnounce(metainfo)
	chPeers := make(chan []ipPort, 1)
	for _, announce := range allAnnounce {
		u, err := url.Parse(announce)
		if err != nil {
			fmt.Printf("parse announce url error: %s\n", err)
			return
		}
		fmt.Println(u)
		if u.Scheme != "http" {
			fmt.Printf("unsupported tracker scheme yet: %s", announce)
			continue
		}

		keepAliveWithTracker(u, metainfo, chPeers)
	}

	for {
		var p []ipPort
		select {
		case p = <-chPeers:
			fmt.Printf("got %d peers\n", len(p))
			peers = uniquePeers(append(peers, p...))
		}
	}

}

func uniquePeers(peers []ipPort) []ipPort {
	keys := make(map[uint64]bool)
	list := []ipPort{}
	for _, entry := range peers {
		e := uint64(entry.Port)<<32 + uint64(entry.IP)
		if _, value := keys[e]; !value {
			keys[e] = true
			list = append(list, entry)
		}
	}
	return list
}
func keepAliveWithTracker(u *url.URL, metainfo map[string]interface{}, chPeers chan []ipPort) {
	q := NewTrackerRequest(metainfo["info"].(map[string]interface{})).Query()

	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		// handle error
		fmt.Printf("GET %s error: %s\n", u.String(), err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("GET %s read error: %s\n", u.String(), err)
		return
	}
	r, err := Parse(body)
	if err != nil {
		fmt.Printf("GET %s get %s parse error: %s", u.String(), string(body), err)
		return
	}
	res := r.(map[string]interface{})
	if res["failure reason"] != nil {
		fmt.Printf("GET %s failure reason: %s", u.String(), res["failure reason"].(string))
		return
	}
	interval := res["interval"].(int)
	fmt.Printf("%s interval: %d\n", u.String(), interval)

	peers := res["peers"]
	switch t := peers.(type) {
	default:
		fmt.Printf("unexpected type %T\n", t) // %T prints whatever type t has
		return
	case string:
		p, err := compactPeerList(peers.(string))
		if err != nil {
			fmt.Printf("parse compact peer list error: %v\n", err)
			return
		}
		if len(p) != 0 {
			chPeers <- p
		}
	case map[string]interface{}:
		p, err := peerList(peers.(map[string]interface{}))
		if err != nil {
			fmt.Printf("parse compact peer list error: %v\n", err)
			return
		}
		if len(p) != 0 {
			chPeers <- p
		}
	}
	time.Sleep(time.Duration(interval * int(time.Second)))
	keepAliveWithTracker(u, metainfo, chPeers)
}

func compactPeerList(peers string) ([]ipPort, error) {
	b := []byte(peers)
	ret := make([]ipPort, 0)
	if len(b)%6 != 0 {
		return nil, errors.New("compact peers is not 6s")
	}
	for i := 0; i < len(b)/6; i++ {
		ip := int(b[i*6])*0xFFFFFF + int(b[i*6+1])*0xFFFF + int(b[i*6+2])*0xFF + int(b[i*6+3])
		port := int(b[i*6+4])*0xFF + int(b[i*6+5])
		i := ipPort{
			IP:   uint32(ip),
			Port: uint16(port),
		}
		ret = append(ret, i)
	}
	return ret, nil
}

func peerList(peers map[string]interface{}) ([]ipPort, error) {
	ret := make([]ipPort, 0)
	for _, p := range peers {
		port, err := strconv.Atoi(p.(map[string]interface{})["port"].(string))
		if err != nil {
			return ret, err
		}
		i := ipPort{
			IP:   uint32(StringIpToInt(p.(map[string]interface{})["ip"].(string))),
			Port: uint16(port),
		}
		ret = append(ret, i)
	}
	return ret, nil
}
func StringIpToInt(ipstring string) int {
	ipSegs := strings.Split(ipstring, ".")
	var ipInt int = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

func IpIntToString(ipInt int) string {
	ipSegs := make([]string, 4)
	var len int = len(ipSegs)
	buffer := bytes.NewBufferString("")
	for i := 0; i < len; i++ {
		tempInt := ipInt & 0xFF
		ipSegs[len-i-1] = strconv.Itoa(tempInt)
		ipInt = ipInt >> 8
	}
	for i := 0; i < len; i++ {
		buffer.WriteString(ipSegs[i])
		if i < len-1 {
			buffer.WriteString(".")
		}
	}
	return buffer.String()
}
func getAllAnnounce(metainfo map[string]interface{}) (ret []string) {
	if metainfo["announce"] == nil {
		log.Fatal("no announce key")
	}
	if metainfo["announce-list"] != nil {
		ret = make([]string, 0, len(metainfo["announce-list"].([]interface{}))+1)
	} else {
		ret = make([]string, 0, 1)
	}
	ret = append(ret, metainfo["announce"].(string))
	if metainfo["announce-list"] != nil {
		announceList := metainfo["announce-list"].([]interface{})
		for _, announce := range announceList {
			ret = append(ret, announce.([]interface{})[0].(string))
		}
	}
	return unique(ret)
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
func parseBTFile(filename string) (map[string]interface{}, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	vv, err := Parse(dat)
	if err != nil {
		return nil, err
	}
	return vv.(map[string]interface{}), nil
}

func ensureFile(info map[string]interface{}) error {
	if info["length"] == nil {
		return ensureFiles(info)
	}
	return ensureOneFile(info)
}
func ensureFiles(info map[string]interface{}) error {
	r := string(append([]rune(downloadRoot), os.PathSeparator))
	filename := r + info["name"].(string)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := os.Mkdir(filename, 0664)
		if err != nil {
			return err
		}
	}

	for _, file := range info["files"].([]map[string]interface{}) {
		path := file["path"].([]string)
		prefix := string(append([]rune(filename), os.PathSeparator))
		err := ensureFileOneByPathList(prefix, path)
		if err != nil {
			return err
		}
	}

	infoFilename := r + info["name"].(string) + ".btinfo"
	if _, err := os.Stat(infoFilename); os.IsNotExist(err) {
		piecesLen := len([]byte(info["pieces"].(string))) / infoHashSize
		bitmapSize := piecesLen / 8
		if piecesLen%8 != 0 {
			bitmapSize++
		}
		b := make([]byte, bitmapSize)
		err := ioutil.WriteFile(infoFilename, b, 0664)
		if err != nil {
			return err
		}
	}
	return nil
}
func ensureFileOneByPathList(rootDir string, pathList []string) error {
	r := []rune(rootDir)
	for i, path := range pathList {
		if i == len(pathList)-1 {
			r = append(r, []rune(path)...)
			filename := string(r)
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				b := make([]byte, 0)
				err := ioutil.WriteFile(filename, b, 0664)
				if err != nil {
					return err
				}
			}
		} else {
			r = append(r, []rune(path)...)
			dir := string(r)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err := os.Mkdir(dir, 0664)
				if err != nil {
					return err
				}
			}
			r = append(r, os.PathSeparator)
		}
	}
	return nil
}
func ensureOneFile(info map[string]interface{}) error {
	r := string(append([]rune(downloadRoot), os.PathSeparator))
	filename := r + info["name"].(string)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		b := make([]byte, 0)
		err := ioutil.WriteFile(filename, b, 0664)
		if err != nil {
			return err
		}
	}

	infoFilename := r + info["name"].(string) + ".btinfo"
	if _, err := os.Stat(infoFilename); os.IsNotExist(err) {
		piecesLen := len([]byte(info["pieces"].(string))) / infoHashSize
		bitmapSize := piecesLen / 8
		if piecesLen%8 != 0 {
			bitmapSize++
		}
		b := make([]byte, bitmapSize)
		err := ioutil.WriteFile(infoFilename, b, 0664)
		if err != nil {
			return err
		}
	}
	return nil
}

func udpTracker(address string, metainfo map[string]interface{}) error {
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
	req := NewTrackerRequest(metainfo)
	announceRequest(conn, transactionID, connectionID, req)

	// IPv4 announce response
	announceResponse(conn, transactionID)
	return nil
}
func announceResponse(conn *net.UDPConn, transactionID uint32) (*TrackerResponse, error) {

	action, err := read4Byte(conn)
	if err != nil {
		return nil, err
	}
	// action          1 // announce
	if action != 1 {
		log.Fatal("action not 1")
	}

	t, err := read4Byte(conn)
	if err != nil {
		return nil, err
	}
	if t != transactionID {
		log.Fatal("transactionID mismatch")
	}

	interval, err := read4Byte(conn)
	fmt.Printf("interval %d\n", interval)

	leechers, err := read4Byte(conn)
	fmt.Printf("leechers %d\n", leechers)

	seeders, err := read4Byte(conn)
	fmt.Printf("seeders %d\n", seeders)

	lst := make([]ipPort, 0)
	for {
		ip, err := read4Byte(conn)
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
func read4Byte(conn *net.UDPConn) (uint32, error) {
	b := make([]byte, 4)
	n, err := conn.Read(b)
	if err != nil {
		return 0, err
	}
	if n != 4 {
		log.Fatal("no 4")
	}
	var i uint32
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		log.Fatal(err)
	}
	return i, nil
}
func read2Byte(conn *net.UDPConn) (uint16, error) {
	b := make([]byte, 2)
	n, err := conn.Read(b)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		log.Fatal("no 2")
	}
	var i uint16
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		log.Fatal(err)
	}
	return i, nil
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
func NewTrackerRequest(info map[string]interface{}) *TrackerRequest {
	r := TrackerRequest{
		InfoHash:   infoHash(info),
		PeerID:     peerID(),
		Port:       availablePort(),
		Uploaded:   0,
		Downloaded: 0,
		Left:       left(info),
		// Key        uint32
		NumWant: -1,
	}
	return &r
}
func (r *TrackerRequest) Query() url.Values {
	v := url.Values{}
	v.Set("info_hash", string(r.InfoHash[:]))
	v.Set("peer_id", string(r.PeerID[:]))
	v.Set("port", strconv.Itoa(int(r.Port)))
	v.Set("uploaded", strconv.Itoa(int(r.Uploaded)))
	v.Set("downloaded", strconv.Itoa(int(r.Downloaded)))
	v.Set("left", strconv.Itoa(int(r.Left)))
	return v
}

func left(info map[string]interface{}) uint64 {
	r := string(append([]rune(downloadRoot), os.PathSeparator))
	infoFilename := r + info["name"].(string) + ".btinfo"
	b, err := ioutil.ReadFile(infoFilename)
	if err != nil {
		log.Fatal(err)
	}

	sum := uint64(0)
	for _, ch := range b {
		sum += uint64(byteTable[ch])
	}
	return sum
}

func availablePort() uint16 {
	for port := 6881; port <= 6889; port++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		defer ln.Close()
		if err == nil {
			return uint16(port)
		}
	}
	return 0
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

func peerID() [peerIDSize]byte {
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
func infoHash(info map[string]interface{}) [infoHashSize]byte {
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
