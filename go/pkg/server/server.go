package server

import (
	"log"
	"net"
	"strings"
)

// TODO: add configuration options
const listeningAddr = ":5000"
const maxPacketSize = 1024

const udpNetworkName = "udp"

var conn *net.UDPConn
var addrStore addressStore

func init() {
	addr, err := net.ResolveUDPAddr(udpNetworkName, listeningAddr)
	if err != nil {
		panic(err)
	}

	conn, err = net.ListenUDP(udpNetworkName, addr)
	if err != nil {
		panic(err)
	}

	addrStore = domainAddrMap{make(map[string][]string)}
}

func ListenAndServe() {
	buffer := make([]byte, 2*maxPacketSize)

	for {
		handleConnection(buffer)
	}
}

func handleConnection(buffer []byte) {
	// TODO: put this in goroutines later
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		// TODO: change logging option later (uber zap logging with logr)
		log.Printf("error: listening address %s: %v", conn.RemoteAddr().String(), err)
		return
	}
	if n > maxPacketSize {
		log.Printf("warn: package by %s: message with %d bytes exceeded max packet size of %d bytes", addr.String(), n, maxPacketSize)
		return
	}

	id := string(buffer[:n])

	remoteAddrs, err := addrStore.ProcessAddress(id, addr.String())
	if err != nil {
		log.Println(err)
		return
	}

	payload := strings.Join(remoteAddrs, ",")
	_, err = conn.WriteToUDP([]byte(payload), addr)
	if err != nil {
		log.Printf("error: writing to socket listening on %s: %v", conn.RemoteAddr().String(), err)
		return
	}
}