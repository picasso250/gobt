package gobt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

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

	return nil
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
