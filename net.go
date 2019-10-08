package gobt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
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
func availablePort() (net.Listener, uint16, error) {
	for port := 6881; port <= 6889; port++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return nil, 0, err
		}
		return ln, uint16(port), nil

	}
	return nil, 0, errors.New("no port available")
}

// StringIPToInt string IP to int
func StringIPToInt(ipstring string) uint32 {
	ipSegs := strings.Split(ipstring, ".")
	var ipInt uint32 = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | uint32(tempInt)
		pos -= 8
	}
	return ipInt
}

// IPIntToString IP int to string
func IPIntToString(ipInt int) string {
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
