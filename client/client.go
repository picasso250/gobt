package main

import (
	"github.com/picasso250/gobt"
)

func main() {
	// dat, err := ioutil.ReadFile("..\\b.torrent")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// v, err := gobt.Parse(dat)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// gobt.PrintMetainfo(v.(map[string]interface{}))
	gobt.Download("Mutant.Year.Zero.Road.to.Eden.Seed.of.Evil.torrent")
}
