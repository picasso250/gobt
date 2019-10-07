package gobt

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
// todo should be int64
type MetainfoInfo struct {
	Name        string
	PieceLength int    // piece length maps to the number of bytes in each piece the file is split into
	Pieces      []byte // pieces maps to a string whose length is a multiple of 20
	Length      int    // There is also a key length or a key files, but not both or neither
	Files       []File
	OriginData  map[string]interface{}
}

// NewMetainfoInfoFromMap builds a map
func NewMetainfoInfoFromMap(m map[string]interface{}) *MetainfoInfo {
	mi := MetainfoInfo{
		Name:        m["name"].(string),
		PieceLength: m["piece length"].(int),
		Pieces:      []byte(m["pieces"].(string)),
		Length:      m["length"].(int),
		OriginData:  m,
	}
	if m["files"] != nil {
		for _, f := range m["files"].([]interface{}) {
			mi.Files = append(mi.Files, NewFileFromMap(f.(map[string]interface{})))
		}

	}

	return &mi
}
func (info *MetainfoInfo) filename() string {
	return pathBuild(downloadRoot, info.Name)
}
func (info *MetainfoInfo) infofilename() string {
	return pathBuild(downloadRoot, info.Name+".btinfo")
}
func (info *MetainfoInfo) bitfield() (bitfield, error) {
	return bitfieldFromFile(info.infofilename())
}

type peerID [peerIDSize]byte
type hash [hashSize]byte

func stringSlice(a []interface{}) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = v.(string)
	}
	return b
}
