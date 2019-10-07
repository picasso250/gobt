package gobt

import (
	"encoding/binary"
	"log"
	"net"
)

func writeInteger(conn net.Conn, i interface{}) error {
	// All later integers sent in the protocol are encoded as four bytes big-endian.
	return binary.Write(conn, binary.BigEndian, i)
}
func read4Byte(conn *net.UDPConn) (uint32, error) {
	var i uint32
	err := binary.Read(conn, binary.BigEndian, &i)
	if err != nil {
		log.Fatal(err)
	}
	return i, nil
}
func read2Byte(conn *net.UDPConn) (uint16, error) {
	var i uint16
	err := binary.Read(conn, binary.BigEndian, &i)
	if err != nil {
		log.Fatal(err)
	}
	return i, nil
}
