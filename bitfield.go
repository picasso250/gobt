package gobt

import (
	"io/ioutil"
	"log"
	"sync"
)

type bitfield struct {
	lock    *sync.RWMutex
	bitData []byte
}

func (b *bitfield) BitData() []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.bitData
}
func (b *bitfield) SetBitData(bitData []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.bitData = bitData
}
func (b *bitfield) Len() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return len(b.bitData)
}
func (b *bitfield) Bit(i int) int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	bi := i / 8
	ii := i % 8
	bv := b.bitData[bi]
	bv >>= (7 - ii)
	return int(bv) & 1
}
func (b bitfield) SetBit(i int, v byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if v|1 != v {
		log.Fatal("bit must be 0 or 1")
	}
	bi := i / 8
	ii := i % 8
	v <<= (7 - ii)
	b.bitData[bi] |= v
}

// size: count of pieces
func allZeroBitField(bitCount int) *bitfield {
	bitmapSize := bitCount / 8
	if bitCount%8 != 0 {
		bitmapSize++
	}
	return &bitfield{
		new(sync.RWMutex),
		make([]byte, bitmapSize),
	}
}
func allZeroBitFieldByte(byteCount int) *bitfield {
	return &bitfield{
		new(sync.RWMutex),
		make([]byte, byteCount),
	}
}
func bitfieldFromFile(filename string) (*bitfield, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	bf := &bitfield{
		new(sync.RWMutex),
		b,
	}
	return bf, nil
}
func (b *bitfield) ToFile(filename string) error {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return ioutil.WriteFile(filename, b.bitData, 0664)
}
func (b *bitfield) left() uint64 {
	sum := uint64(0)
	for _, ch := range b.bitData {
		sum += uint64(byteTable[ch])
	}
	return sum
}
