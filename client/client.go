package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("z", os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		log.Fatal(err)
	}
	o, err := f.Seek(188, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("new offset: %d\n", o)
	n, err := f.Write([]byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("write: %d\n", n)
}
