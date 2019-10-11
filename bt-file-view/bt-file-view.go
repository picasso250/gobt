package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/picasso250/gobt"
)

var onlyName bool

func main() {

	flag.BoolVar(&onlyName, "only-name", false, "only print the name")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <bt_file>\n", os.Args[0])
		os.Exit(1)
	}

	filename := flag.Arg(0)

	fi, err := os.Lstat(filename)
	if err != nil {
		log.Fatal(err)
	}

	switch mode := fi.Mode(); {
	default:
		fmt.Fprintf(os.Stderr, "unknown file type %d\n", mode)
		os.Exit(1)
	case mode.IsRegular():
		err := doFile(filename, onlyName)
		if err != nil {
			log.Fatal(err)
		}
	case mode.IsDir():
		err := doDir(filename, onlyName)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func doFile(filename string, onlyName bool) error {
	mi, err := gobt.NewMetainfoFromFile(filename)
	if err != nil {
		return err
	}
	if onlyName {
		fmt.Println(mi.Info.Name)
	} else {
		fmt.Println(mi)
	}
	return nil
}
func doDir(dirname string, onlyName bool) error {
	f, err := os.Open(dirname)
	if err != nil {
		return err
	}
	defer f.Close()
	files, err := f.Readdirnames(10000)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(dirname, file)
		if !strings.HasSuffix(path, ".torrent") {
			continue
		}
		fmt.Printf("%s: ", file)
		doFile(path, onlyName)
	}
	return nil
}
