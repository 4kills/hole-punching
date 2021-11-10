package main

import "net"

const network = "udp"
const wellKownAddr = ":5000"

func main() {
	laddr, _ := net.ResolveUDPAddr(network, wellKownAddr)
	conn, _ := net.ListenUDP(network, laddr)

	b := make([]byte, 0xff)

	for {
		_, fst, _ := conn.ReadFromUDP(b)
		_, snd, _ := conn.ReadFromUDP(b)

		conn.WriteToUDP([]byte(snd.String()), fst)
		conn.WriteToUDP([]byte(fst.String()), snd)
	}
}