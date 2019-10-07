package main

import (
	"fmt"
	"log"

	"github.com/picasso250/gobt"
)

func main() {
	i, err := gobt.Parse([]byte("i11e"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("hello %d\n", i)
}
