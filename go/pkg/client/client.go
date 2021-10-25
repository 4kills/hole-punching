package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const network = "udp"

var timeout = time.Second * 7

func Connect(id []byte, wellKnown string, expected int) ([]*net.UDPAddr, error) {
	socket, err := net.ListenUDP(network, &net.UDPAddr{})
	if err != nil {
		return nil, err
	}

	wellKnownUPD, err := net.ResolveUDPAddr(network, wellKnown)
	if err != nil {
		return nil, err
	}

	readBuffer := make([]byte, 1<<16)
	var remConns []*net.UDPAddr

	// TODO add max tries/deadline/timeout
	for {
		fmt.Println("Try") // TODO: remove this later
		_, err = socket.WriteToUDP(id, wellKnownUPD)
		if err != nil {
			return nil, err
		}

		socket.SetReadDeadline(time.Now().Add(timeout))
		n, _, err := socket.ReadFromUDP(readBuffer)
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		} else if err != nil {
			return nil, err
		}
		if n == 0 { // possibly keep alive packet
			continue
		}

		remConns, err = parse(readBuffer[:n])
		if err != nil {
			return nil, err
		}

		if len(remConns) == expected {
			break
		}
	}

	wg := &sync.WaitGroup{}
	for _, c := range remConns {
		wg.Add(1)
		go connectIndividual(c, socket, wg)
	}

	wg.Wait()

	return remConns, nil
}

func connectIndividual(conn *net.UDPAddr, socket *net.UDPConn, wg *sync.WaitGroup) {
	readBuffer := make([]byte, 0xffff) // TODO: adjust size
	// TODO: deadline wont work across multiple goroutines simultaneously
	socket.SetReadDeadline(time.Now().Add(timeout))
	fmt.Println("Try sending packet to " + conn.String()) // TODO: remove this

	msgChan := make(chan string, 1)
	msgChan <- "SYN"

	go func() {
		var ok bool
		var msg string
		for {
			select {
			case msg, ok = <-msgChan:
				if !ok {
					return
				}
			default:
				_, err := socket.WriteToUDP([]byte(msg), conn)
				if err != nil {
					panic(err) // TODO: change this behavior to sth sensical
				}

				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for {
		n, _, err := socket.ReadFromUDP(readBuffer)
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue // TODO: change this
		} else if err != nil {
			panic(err)
		}
		if n == 0 { // keep alive packet
			continue
		}

		str := string(readBuffer[:n])
		fmt.Println(str)

		if str == "SYN" {
			msgChan <- "ACK"
		} else if str == "ACK" {
			close(msgChan)
			break
		} else {
			log.Printf("warn: illegal request by %s with message: %s", conn.String(), str)
		}

	}

	wg.Done()
}

func parse(content []byte) ([]*net.UDPAddr, error) {
	rawAddrs := strings.Split(string(content), ",")
	ret := make([]*net.UDPAddr, len(rawAddrs))

	for i, v := range rawAddrs {
		addr, err := net.ResolveUDPAddr(network, v)
		if err != nil {
			return ret, err
		}
		ret[i] = addr
	}

	return ret, nil
}
