package server

import (
	"log"
	"net"
)

const listeningAddr = ":5000"
const maxPacketSize = 1024

const udpNetworkName = "udp"

var conn *net.UDPConn

var identifiers map[string][]string

func init() {
	addr, err := net.ResolveUDPAddr(udpNetworkName, listeningAddr)
	if err != nil {
		panic(err)
	}

	conn, err = net.ListenUDP(udpNetworkName, addr)
	if err != nil {
		panic(err)
	}

}

func listen() {
	buffer := make([]byte, 2 * maxPacketSize)

	for {
		// TODO: put this in goroutines later
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// TODO: change logging option later (uber zap logging with logr)
			log.Printf("error: ip %v; port %v: %v", addr.IP, addr.Port, err)
		}
		if n > maxPacketSize {
			log.Printf("warn: ip %v; port %v: message with %d bytes exceeded max packet size (%d bytes)", addr.IP, addr.Port, n, maxPacketSize)
		}

		id := string(buffer)

	}
}

func fetchAddr(id, except string) {
	
}