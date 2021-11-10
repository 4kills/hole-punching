package main

import (
	"github.com/4kills/hole-punching/go/pkg/client"
	"log"
)

func main() {
	c, err := client.New("4kills.net:5000")
	if err != nil {
		panic(err)
	}

	_, _, err = c.Connect([]byte("domain"), 1)
	if err != nil {
		log.Println(err)
	}
}