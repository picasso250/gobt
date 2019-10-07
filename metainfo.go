package gobt

// Metainfo Metainfo files (also known as .torrent files)
type Metainfo struct {
	Announce     string
	AnnounceList []string
	Info         *MetainfoInfo
	InfoHash     [infoHashSize]byte
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
	PieceLength int
	Pieces      []byte
	Length      int
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

// File file
// todo should be int64
type File struct {
	Length int
	Path   []string
}

// NewFileFromMap builds a File
func NewFileFromMap(m map[string]interface{}) File {
	return File{
		Length: m["length"].(int),
		Path:   stringSlice(m["path"].([]interface{})),
	}
}

type peerID [peerIDSize]byte

func stringSlice(a []interface{}) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = v.(string)
	}
	return b
}
