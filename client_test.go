package gobt

import (
	"log"
	"testing"
)

func TestDownload(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	Download("Mutant.Year.Zero.Road.to.Eden.Seed.of.Evil.torrent")
	t.Errorf("not implemented")
}
