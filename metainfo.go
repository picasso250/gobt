package gobt

import (
	"errors"
	"io/ioutil"
)

// Metainfo Metainfo files (also known as .torrent files)
type Metainfo struct {
	Announce     string
	AnnounceList []string
	Info         *MetainfoInfo
	InfoHash     hash
	OriginData   map[string]interface{}
}

// NewMetainfoFromMap builds a Metainfo
func NewMetainfoFromMap(m map[string]interface{}) *Metainfo {
	info := m["info"].(map[string]interface{})
	mi := Metainfo{
		Announce:   string(m["announce"].([]byte)),
		Info:       NewMetainfoInfoFromMap(info),
		InfoHash:   infoHash(info),
		OriginData: m,
	}
	if m["announce-list"] != nil {
		for _, a := range flat(m["announce-list"].([]interface{})) {
			mi.AnnounceList = append(mi.AnnounceList, string(a.([]byte)))
		}
	}
	return &mi
}

// NewMetainfoFromFile read file and return metainfo
func NewMetainfoFromFile(filename string) (*Metainfo, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	vv, err := Parse(dat)
	if err != nil {
		return nil, err
	}
	return NewMetainfoFromMap(vv.(map[string]interface{})), nil
}
func (m *Metainfo) String() string {
	return valueToString(m.OriginData)
}
func flat(a []interface{}) []interface{} {
	ret := make([]interface{}, 0, len(a))
	for _, v := range a {
		if v, ok := v.([]interface{}); ok {
			for _, v2 := range v {
				ret = append(ret, v2)
			}
		}
	}
	return ret
}

// MetainfoInfo metainfo[info]
type MetainfoInfo struct {
	Name        string
	PieceLength int    // piece length maps to the number of bytes in each piece the file is split into
	Pieces      []byte // pieces maps to a string whose length is a multiple of 20
	Length      int64  // There is also a key length or a key files, but not both or neither
	Files       []File // But we always assign Length as total length for convenience
	OriginData  map[string]interface{}
}

// NewMetainfoInfoFromMap builds a map
func NewMetainfoInfoFromMap(m map[string]interface{}) *MetainfoInfo {
	mi := MetainfoInfo{
		Name:        string(m["name"].([]byte)),
		PieceLength: int(m["piece length"].(int64)),
		Pieces:      m["pieces"].([]byte),
		OriginData:  m,
	}
	if m["length"] != nil {
		mi.Length = m["length"].(int64)
	}
	if m["files"] != nil {
		for _, f := range m["files"].([]interface{}) {
			mi.Files = append(mi.Files, NewFileFromMap(f.(map[string]interface{})))
		}
	}

	return &mi
}

func (info *MetainfoInfo) piecesCount() int {
	return len([]byte(info.Pieces)) / hashSize
}
func (info *MetainfoInfo) filename() string {
	return buildPath(DownloadRoot, info.Name)
}
func (info *MetainfoInfo) infoFilename() string {
	return info.filename() + ".btinfo"
}
func (info *MetainfoInfo) bitfield() (*bitfield, error) {
	return bitfieldFromFile(info.infoFilename())
}

type peerID [peerIDSize]byte

func peerIDFromBytes(b []byte) (peerID, error) {
	var pid peerID
	if len(b) != peerIDSize {
		return pid, errors.New("peer id is not length 20")
	}
	copy(pid[:], b)
	return pid, nil
}

type hash [hashSize]byte

func stringSlice(a []interface{}) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = string(v.([]byte))
	}
	return b
}
