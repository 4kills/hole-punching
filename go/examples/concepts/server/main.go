package main

import (
	"log"
	"net"
	"os"
)

const network = "udp"
var wellKownAddr = ":5001"

func main() {
	if len(os.Args) > 1 {
		wellKownAddr = os.Args[1]
	}

	laddr, err := net.ResolveUDPAddr(network, wellKownAddr)
	if err != nil {
		log.Println(err)
	}
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		log.Println(err)
	}

	b := make([]byte, 0xff)

	for {
		_, fst, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println(err)
		}
		_, snd, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println(err)
		}

		conn.WriteToUDP([]byte(snd.String()), fst)
		conn.WriteToUDP([]byte(fst.String()), snd)
	}
}