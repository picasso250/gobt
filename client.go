package gobt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
)

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

func udpTracker(address string) error {
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
	var transactionID uint32 = uint32(rand.Int31())
	b, err := connectRequest(transactionID)
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

	buffer := make([]byte, 1024)
	n, err = conn.Read(buffer)
	if err != nil {
		return err
	}
	fmt.Printf("read %d bytes\n", n)
	if n != 16 {
		log.Fatal("not 16 read")
	}
	action := uint32(0)
	err = binary.Read(conn, binary.BigEndian, &action)
	if err != nil {
		return err
	}
	if action != 0 {
		log.Fatal("action not 0")
	}
	transID := uint32(0)
	err = binary.Read(conn, binary.BigEndian, &transID)
	if err != nil {
		return err
	}
	if transID != transactionID {
		return errors.New("transaction_id not equal")
	}
	var connectionID uint64
	err = binary.Read(conn, binary.BigEndian, &connectionID)
	if err != nil {
		return err
	}
	return nil
}

func connectRequest(transactionID uint32) ([]byte, error) {
	var protocolID uint64 = 0x41727101980 // magic constant
	var action uint32 = 0
	var err error
	buf := bytes.NewBuffer(make([]byte, 0, 16))
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
