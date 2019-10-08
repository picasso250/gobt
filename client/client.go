package main

import (
	"io/ioutil"
	"log"

	"github.com/picasso250/gobt"
)

func main() {
	dat, err := ioutil.ReadFile("..\\b.torrent")
	if err != nil {
		log.Fatal(err)
	}
	v, err := gobt.Parse(dat)
	if err != nil {
		log.Fatal(err)
	}

	gobt.PrintMetainfo(v.(map[string]interface{}))
}
