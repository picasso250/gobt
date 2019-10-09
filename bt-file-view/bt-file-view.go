package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/picasso250/gobt"
)

func main() {

	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <bt_file>\n", os.Args[0])
		return
	}

	filename := flag.Arg(0)
	mi, err := gobt.NewMetainfoFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mi)
}
