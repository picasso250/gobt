package gobt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

// non-keepalive messages start with a single byte which gives their type
const (
	typeChoke = iota
	typeUnchoke
	typeInterested
	typeNotInterested
	typeHave
	typeBitfield
	typeRequest
	typePiece
	typeCancel
)

const requestLength = uint32(1 << 14) // All current implementations use 2^14 (16 kiB)

func doPeer(ipt ipPort, metainfo *Metainfo) {
	conn, err := net.Dial("tcp4", ipt.String())
	if err != nil {
		fmt.Printf("dial tcp %s error: %s\n", ipt.String(), err)
		return
	}
	defer conn.Close()
	err = handshake(conn, ipt, metainfo)
	if err != nil {
		fmt.Printf("%s handshake error: %s", ipt, err)
		return
	}

	heartBeatChan := make(chan int)
	go heartBeat(conn, heartBeatChan)
	defer func() {
		// inform heart beat to stop
		heartBeatChan <- 1
	}()

	err = peerMessages(conn, metainfo.Info)
	if err != nil {
		fmt.Printf("peer messages error: %s\n", err)
		return
	}

}
func peerMessages(conn net.Conn, info *MetainfoInfo) error {
	msg, err := buildpeerMessageBitfield(info)
	if err != nil {
		return err
	}
	err = sendMessage(conn, msg)
	if err != nil {
		return err
	}

	err = alternatingStream(conn)
	if err != nil {
		return err
	}
}
func alternatingStream(conn net.Conn, info *MetainfoInfo) error {
	// randomly pick a pieace to download
	bf, err := bitfieldFromFile(info.filename())
	if err != nil {
		return err
	}

	index := 0
	start := rand.Intn(info.Length)
	cnt := info.piecesCount()
	for i := 0; i < cnt; i++ {
		ii := (start + i) % cnt
		if bf.Bit(ii) == 0 {
			index = ii
		}
	}

	sendMessage(conn, requestMessage())
}

var pieceInnerIndex int32 = 0

func requestMessage(index int32) *bytes.Buffer {
	b := new(bytes.Buffer)

	err := writeInteger(b, index)
	if err != nil {
		log.Fatal(err)
	}

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

func handshake(conn net.Conn, ipt ipPort, metainfo *Metainfo) error {
	// The handshake starts with character ninteen (decimal) followed by the string 'BitTorrent protocol'
	err := protocol(conn)
	if err != nil {
		fmt.Printf("%s protocol error: %s", ipt, err)
		return err
	}

	err = reservedBytes(conn)
	if err != nil {
		fmt.Printf("%s reserved bytes error: %s", ipt, err)
		return err
	}

	err = exchangeSha1Hash(conn, metainfo.InfoHash[:])
	if err != nil {
		fmt.Printf("%s exchange hash error: %s", ipt, err)
		return err
	}

	err = exchangePeerID(conn, myPeerID[:])
	if err != nil {
		fmt.Printf("%s exchange peer id error: %s", ipt, err)
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

	br := make([]byte, infoHashSize)
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
