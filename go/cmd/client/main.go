package main

import (
	"bufio"
	"fmt"
	"github.com/4kills/hole-punching/go/pkg/client"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type chat struct {
	addrs []*net.UDPAddr
	socket *net.UDPConn

	keepAlivePeriod time.Duration
}

func initConnection() (chat, error) {
	usage := "Usage:\nclient <rendezvous> <id> <num_peers>."
	if len(os.Args) < 4 {
		return chat{}, fmt.Errorf("not enough arguments provided. Provided: %v. %s", os.Args[1:], usage)
	}

	rendezvous := os.Args[1]
	id := os.Args[2]
	numPeers, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return chat{}, fmt.Errorf("cannot convert %s to integer. num_peers not provided. %s", os.Args[3], usage)
	}

	c, err := client.New(rendezvous)
	if err != nil {
		return chat{}, err
	}

	addrs, s, err := c.Connect([]byte(id), numPeers)
	return chat{addrs: addrs, socket: s, keepAlivePeriod: 5 * time.Second}, err
}

func main() {
	c, err := initConnection()
	if err != nil {
		panic(err)
	}

	go c.keepAlive()

	go c.receive()

	c.send()
}

func (c chat) receive() {
	m := map[string]int{}
	for i, addr := range c.addrs {
		m[addr.String()] = i
	}

	b := make([]byte, 0xffff)

	for {
		n, p, _ := c.socket.ReadFromUDP(b)
		if n == 0 {
			continue
		}

		i, ok := m[p.String()]
		from := fmt.Sprintf("Peer %d:", i)
		if !ok {
			from = "Peer unknown:"
		}

		log.Println(from, string(b[:n]))
	}
}

func (c chat) send() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		msg := scanner.Text()
		for _, addr := range c.addrs {
			c.socket.WriteToUDP([]byte(msg), addr)
		}

		log.Println("Me:", msg)
	}
}

func (c chat) keepAlive() {
	for {
		select {
		case <-time.After(c.keepAlivePeriod):
			for _, addr := range c.addrs {
				c.socket.WriteToUDP([]byte{}, addr)
			}
		}
	}
}