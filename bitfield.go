package gobt

import "log"

type bitfield []byte

func (b bitfield) Bit(i int) int {
	bi := i / 8
	ii := i % 8
	bv := b[bi]
	bv >>= (7 - ii)
	return int(bv) & 1
}
func (b bitfield) SetBit(i int, v byte) {
	if v|1 != v {
		log.Fatal("bit must be 0 or 1")
	}
	bi := i / 8
	ii := i % 8
	v <<= (7 - ii)
	b[bi] |= v
}
func allZeroBitField(size int) bitfield {
	bitmapSize := size / 8
	if size%8 != 0 {
		bitmapSize++
	}
	return make(bitfield, bitmapSize)
}
