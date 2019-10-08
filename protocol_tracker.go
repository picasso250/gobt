package gobt

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

// TrackerRequest Tracker GET requests
type TrackerRequest struct {
	InfoHash   hash
	PeerID     peerID
	IP         uint32
	Port       uint16
	Uploaded   uint64
	Downloaded uint64
	Left       uint64
	Event      uint32
	Key        uint32
	NumWant    int32
}

// TrackerResponse IPv4 announce response
type TrackerResponse struct {
	Action        uint32
	TransactionID uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32
	IPPort        []ipPort
}

// NewTrackerRequest new a tracker request with current bt file
func NewTrackerRequest(mi *Metainfo, port uint16) *TrackerRequest {
	r := TrackerRequest{
		InfoHash:   mi.InfoHash,
		PeerID:     myPeerID,
		Port:       port,
		Uploaded:   0,
		Downloaded: 0,
		Left:       gBitField.left() * uint64(mi.Info.PieceLength),
		// Key        uint32
		// NumWant: -1,
	}
	return &r
}

// Query return http query
func (r *TrackerRequest) Query() url.Values {
	v := url.Values{}
	v.Set("info_hash", string(r.InfoHash[:]))
	v.Set("peer_id", string(r.PeerID[:]))
	v.Set("port", strconv.Itoa(int(r.Port)))
	v.Set("uploaded", strconv.Itoa(int(r.Uploaded)))
	v.Set("downloaded", strconv.Itoa(int(r.Downloaded)))
	v.Set("left", strconv.Itoa(int(r.Left)))
	return v
}

var trackerMap map[string]bool // state of tracker
var trackerMapMutex sync.RWMutex

func trackerProtocol(metainfo *Metainfo, port uint16) {
	allAnnounce := getAllAnnounce(metainfo)
	for _, announce := range allAnnounce {
		u, err := url.Parse(announce)
		if err != nil {
			fmt.Printf("parse announce url error: %s\n", err)
			return
		}
		// fmt.Println(u)
		if u.Scheme != "http" {
			fmt.Printf("unsupported tracker scheme yet: %s\n", announce)
			continue
		}

		go keepAliveWithTracker(u, metainfo, port)
	}
}
func keepAliveWithTracker(u *url.URL, metainfo *Metainfo, port uint16) {
	q := NewTrackerRequest(metainfo, port).Query()
	u.RawQuery = q.Encode()
	var body []byte
	if doNotBotherTracker { // for debug
		debugRoot := ".debug"
		cacheFile := buildPath(debugRoot, (u.Hostname()))
		if _, err := os.Stat(cacheFile); err != nil {
			errFile := buildPath(debugRoot, (u.Hostname())+".error")
			if _, err := os.Stat(errFile); err == nil {
				errBytes, err := ioutil.ReadFile(errFile)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("last error of %s: %s\n", u.String(), string(errBytes))
				return
			}
			resp, err := http.Get(u.String())
			if err != nil {
				fmt.Printf("GET %s error: %s\n", u.String(), err)
				err = ioutil.WriteFile(errFile, []byte(err.Error()), 0664)
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("GET %s read error: %s\n", u.String(), err)
				return
			}
			err = ioutil.WriteFile(cacheFile, body, 0664)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("cache hit %s\n", u.String())
			body, err = ioutil.ReadFile(cacheFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		resp, err := http.Get(u.String())
		if err != nil {
			fmt.Printf("GET %s error: %s\n", u.String(), err)
			return
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("GET %s read error: %s\n", u.String(), err)
			return
		}
	}

	r, err := Parse(body)
	if err != nil {
		fmt.Printf("GET %s get %s parse error: %s\n", u.String(), string(body), err)
		return
	}
	res := r.(map[string]interface{})
	if res["failure reason"] != nil {
		fmt.Printf("GET %s failure reason: %s\n", u.String(), res["failure reason"].([]byte))
		return
	}
	interval := res["interval"].(int)
	fmt.Printf("interval: %d\t(%s)\n", interval, u.String())

	peers := res["peers"]
	pl := make([]*peer, 0)
	piecesCount := metainfo.Info.piecesCount()
	switch t := peers.(type) {
	default:
		fmt.Printf("unexpected type %T\n", t) // %T prints whatever type t has
		return
	case string:
		pl, err = compactPeerList(peers.([]byte), piecesCount)
		if err != nil {
			fmt.Printf("parse compact peer list error: %v\n", err)
			return
		}
	case map[string]interface{}:
		pl, err = peerList(peers.(map[string]interface{}), piecesCount)
		if err != nil {
			fmt.Printf("parse compact peer list error: %v\n", err)
			return
		}
	}

	// todo limit the number of peers
	if len(pl) != 0 {
		for _, pp := range pl {
			gPeersToStart <- pp
		}
	}

	time.Sleep(time.Duration(interval) * (time.Second))
	go keepAliveWithTracker(u, metainfo, port)
}
