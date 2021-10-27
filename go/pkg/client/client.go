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

type client struct {
	Timeout time.Duration

	MediatorServerRetryPeriod time.Duration
	Socket                    *net.UDPConn

	wellKnownHost         *net.UDPAddr
}

func New(wellKnownHost string) (client, error) {
	c := client{
		Timeout:                   10 * time.Second,
		MediatorServerRetryPeriod: 100 * time.Millisecond,
	}

	s, err := net.ListenUDP(network, &net.UDPAddr{})
	if err != nil {
		return c, err
	}

	wellKnownUDP, err := net.ResolveUDPAddr(network, wellKnownHost)
	if err != nil {
		return c, err
	}

	c.wellKnownHost = wellKnownUDP
	c.Socket = s
	return c, err
}

func (c client) Connect(id []byte, expected int) ([]*net.UDPAddr, *net.UDPConn , error) {
	err := c.Socket.SetReadDeadline(time.Now().Add(c.Timeout))
	if err != nil {
		return nil, nil, err
	}

	remConns, err := c.connectToServer(id, expected)
	if err != nil {
		return nil, nil, err
	}

	wg := &sync.WaitGroup{}
	for _, peer := range remConns {
		wg.Add(1)
		go c.connectIndividual(peer, wg)
	}

	wg.Wait()

	return remConns, c.Socket, nil
}

func (c client) connectToServer(id []byte, expected int) ([]*net.UDPAddr, error)  {
	var remConns []*net.UDPAddr
	readBuffer := make([]byte, 0xffff)

	chanErr := make(chan error, 1)
	defer close(chanErr)

	go func() {
		for {
			select {
			case <- chanErr:
				return
			default:
				_, err := c.Socket.WriteToUDP(id, c.wellKnownHost)
				if err != nil {
					chanErr <- err
					return
				}

				time.Sleep(c.MediatorServerRetryPeriod)
			}
		}
	}()

	for {
		select {
		case err := <- chanErr:
			return nil, err
		default:
			n, inboundAddr, err := c.Socket.ReadFromUDP(readBuffer)
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return nil, fmt.Errorf("%w: timeout after %s: %v", ErrTimeoutDuringServerConnect, c.Timeout.String(), err)
			} else if err != nil {
				return nil, err
			}
			if n == 0 { // possibly keep alive packet
				continue
			}

			if inboundAddr.String() != c.wellKnownHost.String() {
				continue
			}

			remConns, err = parse(readBuffer[:n])
			if err != nil {
				return nil, err
			}

			if len(remConns) == expected {
				return remConns, nil
			}
		}
	}
}

func (c client) connectIndividual(peer *net.UDPAddr, wg *sync.WaitGroup) {
	readBuffer := make([]byte, 0xffff) // TODO: adjust size

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
				_, err := c.Socket.WriteToUDP([]byte(msg), peer)
				if err != nil {
					panic(err) // TODO: change this behavior to sth sensical
				}

				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for {
		// TODO: this will potentially not work because connection could be from s/o else
		n, _, err := c.Socket.ReadFromUDP(readBuffer)
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
			log.Printf("warn: illegal request by %s with message: %s", peer.String(), str)
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
