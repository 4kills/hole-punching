package server

import (
	"net"
	"strings"
)

// TODO: add configuration options
const listeningAddr = ":5000"
const maxPacketSize = 1024

const udpNetworkName = "udp"

var conn *net.UDPConn
var addrStore AddressStore

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
		log.Error(err, "read from udp with remote address: rejecting address", logKeyAddr, addr.String())
		return
	}
	if n > maxPacketSize {
		log.V(1).Info( "package by remote address with messageLength bytes exceeded maxPacketSize: rejecting address.", logKeyAddr, addr.String(), "messageLength", n, "maxPacketSize", maxPacketSize)
		return
	}

	id := string(buffer[:n])

	remoteAddrs, err := addrStore.ProcessAddress(id, addr.String())
	if err != nil {
		log.Error(err, "could not store address: rejecting address", logKeyAddr, addr.String())
		return
	}

	payload := strings.Join(remoteAddrs, ",")
	_, err = conn.WriteToUDP([]byte(payload), addr)
	if err != nil {
		log.Error(err, "writing to remote address ; socket listening on port", logKeyAddr, addr.String(), "port", conn.RemoteAddr().String())
		return
	}

	log.V(1).Info("wrote package to address with payload", logKeyAddr, addr.String(), "payload", payload)
}