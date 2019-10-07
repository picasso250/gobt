package gobt

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
)

func writeInteger(w io.Writer, i interface{}) error {
	// All later integers sent in the protocol are encoded as four bytes big-endian.
	return binary.Write(w, binary.BigEndian, i)
}
func writeIntegers(w io.Writer, is ...interface{}) error {
	// All later integers sent in the protocol are encoded as four bytes big-endian.
	buf := new(bytes.Buffer)
	for _, v := range is {
		err := binary.Write(buf, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}
	return writeAll(w, buf.Bytes())
}
func readUint32(conn io.Reader) (uint32, error) {
	var i uint32
	err := binary.Read(conn, binary.BigEndian, &i)
	if err != nil {
		return 0, err
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
