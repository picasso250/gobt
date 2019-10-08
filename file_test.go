package gobt

import "testing"

func init() {
	DownloadRoot = ".debug"
}

func TestEnsureFile(t *testing.T) {
	mi, err := newMetainfoFromFile("a.txt.torrent")
	if err != nil {
		t.Errorf("parse file error: %v", err)
	}
	_, err = ensureOneFile(mi.Info)
	if err != nil {
		t.Errorf("ensureFile error %s", err)
	}

	_, err = ensureFile(mi.Info)
	if err != nil {
		t.Errorf("ensureFile error %s", err)
	}
}
func TestEnsureFiles(t *testing.T) {
	mi, err := newMetainfoFromFile("b.torrent")
	if err != nil {
		t.Errorf("parse file error: %v", err)
	}

	_, err = ensureFiles(mi.Info)
	if err != nil {
		t.Errorf("ensureFile error %s", err)
	}
}
