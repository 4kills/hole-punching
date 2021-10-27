package client

import (
	"context"
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
	// Timeout sets the duration after which client.Connect will time out and return with an error. If the value is negative,
	// client.Connect will never time out.
	Timeout time.Duration

	MediatorServerRetryPeriod time.Duration
	PeerRetryPeriod			  time.Duration
	Socket                    *net.UDPConn

	wellKnownHost         *net.UDPAddr
}

func New(wellKnownHost string) (client, error) {
	c := client{
		Timeout:                   10 * time.Second,
		MediatorServerRetryPeriod: 100 * time.Millisecond,
		PeerRetryPeriod: 		   100 * time.Millisecond,
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
	if c.Timeout >= 0 {
		err := c.Socket.SetReadDeadline(time.Now().Add(c.Timeout))
		if err != nil {
			return nil, nil, err
		}
	}

	remConns, err := c.connectToServer(id, expected)
	if err != nil {
		return nil, nil, err
	}

	err = c.connectPeers(remConns)

	return remConns, c.Socket, err
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

func (c client) connectPeers(remConns []*net.UDPAddr) error {
	readBuffer := make([]byte, 0xffff)
	remotes := make(map[string]chan string)
	cErr := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer cancel()

	for _, peer := range remConns {
		ch := make(chan string, 1)
		remotes[peer.String()] = ch
		wg.Add(1)
		go c.connectIndividual(peer, ch, cErr, ctx, cancel, wg)
	}

	cWait := make(chan struct{})
	go func() {
		wg.Wait()
		close(cWait)
	}()

	for {
		select {
		case <- cWait:
			return nil
		case err := <- cErr:
			return err
		default:
			n, inbound, err := c.Socket.ReadFromUDP(readBuffer)
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return fmt.Errorf("%w: timeout after %s: %v", ErrTimeoutDuringPeerConnect, c.Timeout.String(), err)
			} else if err != nil {
				return err
			}
			if n == 0 { // keep alive packet
				continue
			}

			remotes[inbound.String()] <- string(readBuffer[:n])
		}
	}
}

func (c client) connectIndividual(peer *net.UDPAddr, ch chan string, cErr chan error, ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	syn := "SYN"
	ack := "ACK"

	msgChan := make(chan string, 1)
	defer close(msgChan)
	msgChan <- syn

	go func() {
		var ok bool
		var msg string
		for {
			select {
			case <- ctx.Done():
				return
			case msg, ok = <-msgChan:
				if !ok {
					return
				}
			default:
				_, err := c.Socket.WriteToUDP([]byte(msg), peer)
				if err != nil {
					cErr <- err
					cancel()
					return
				}

				time.Sleep(c.PeerRetryPeriod)
			}
		}
	}()

	for {
		select {
		case <- ctx.Done():
			return
		case str := <- ch:
			if str == syn {
				msgChan <- ack
			} else if str == ack {
				wg.Done()
				return
			} else {
				// log.Printf("warn: illegal request by %s with message: %s", peer.String(), str)
			}
		}
	}
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
