package gobt

import "errors"

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
		Announce:   m["announce"].(string),
		Info:       NewMetainfoInfoFromMap(info),
		InfoHash:   infoHash(info),
		OriginData: m,
	}
	if m["announce-list"] != nil {
		for _, a := range m["announce-list"].([]interface{}) {
			mi.AnnounceList = append(mi.AnnounceList, a.([]interface{})[0].(string))
		}
	}
	return &mi
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
		Name:        m["name"].(string),
		PieceLength: int(m["piece length"].(int64)),
		Pieces:      []byte(m["pieces"].(string)),
		Length:      m["length"].(int64),
		OriginData:  m,
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
	return pathBuild(downloadRoot, info.Name)
}
func (info *MetainfoInfo) infofilename() string {
	return pathBuild(downloadRoot, info.Name+".btinfo")
}
func (info *MetainfoInfo) bitfield() (*bitfield, error) {
	return bitfieldFromFile(info.infofilename())
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
		b[i] = v.(string)
	}
	return b
}
