package main

import (
	"github.com/4kills/hole-punching/pkg/client"
	"log"
)

func main() {
	_, err := client.Connect([]byte("domain"), "4kills.net:5000", 1)
	if err != nil {
		log.Println(err)
	}
}