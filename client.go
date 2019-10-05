package gobt

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
	"strconv"
)

const infoHashSize = 20
const peerIDSize = 20

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

// Download download BT file
func Download(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	vv, err := Parse(dat)
	if err != nil {
		return err
	}
	v := vv.(map[string]interface{})
	if v["announce"] == nil {
		return errors.New("no announce key")
	}
	announce := v["announce"].(string)
	u, err := url.Parse(announce)
	if err != nil {
		return err
	}
	fmt.Println(u.Scheme)
	fmt.Println(u)
	if u.Scheme != "udp" {
		return errors.New("unsupported scheme")
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
	return nil
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
		Left:       left(),
		// Key        uint32
		NumWant: -1,
	}
	return &r
}

func availablePort() uint16 {
	for port := 6881; port <= 6889; port++ {
		_, err := net.Listen("tcp", ":"+strconv.Itoa(port))
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
