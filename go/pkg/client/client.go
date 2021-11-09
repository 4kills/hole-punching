package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const network = "udp"

type client struct {
	// Timeout sets the duration after which Connect will time out and return with an error. If the value is negative,
	// Connect will never time out.
	Timeout time.Duration

	// MediatorServerRetryPeriod sets the delay between each ping to the server.
	MediatorServerRetryPeriod time.Duration
	// PeerRetryPeriod sets the delay between each packet being sent to a remote
	PeerRetryPeriod			  time.Duration
	// Socket represents the instance (LADDR:LPORT) used to establish the connections. THIS SOCKET HAS TO BE USED
	// FOR FURTHER COMMUNICATION.
	Socket                    *net.UDPConn

	wellKnownHost         *net.UDPAddr
	readDeadline	      time.Time
}

// New returns a new client used to establish peer connections through the wellKnownHost. After connection, you
// will have to extract the client.Socket for further communication.
func New(wellKnownHost string) (client, error) {
	c := client{
		Timeout:                   40 * time.Second,
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

// Connect returns all connected peers (i.e. their respective UDPAddr) as well as the UDPConn used for connection.
// You MUST use this UDPConn for further communication. This is the same as client.Socket.
//
// Connect uses id to identify peers trying to connect through the same id. Expected is the number of peers expected to connect.
// When expected numbers of peers have connected this method returns with a nil error. When not all peers connect in client.Timeout
// an ErrTimeoutDuringPeerConnect (wrapping os.ErrDeadlineExceeded) will be returned. However, the returned UPDAddr might still be of use.
//
// If the client times out during attempting to connect to the server (server not available) or if not expected number of peers have reached
// the server yet, the method will return ErrTimeoutDuringServerConnect (wrapping os.ErrDeadlineExceeded) with []UDPAddr containing all addr so far.
// You may then try to connect with these peers using ConnectPeers.
func (c client) Connect(id []byte, expected int) ([]*net.UDPAddr, *net.UDPConn , error) {
	if c.Timeout >= 0 {
		c.readDeadline = time.Now().Add(c.Timeout)
		err := c.Socket.SetReadDeadline(c.readDeadline)
		if err != nil {
			return nil, nil, err
		}
	}

	remConns, err := c.connectToServer(id, expected)
	if err != nil {
		return remConns, c.Socket, err
	}

	err = c.ConnectPeers(remConns)

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

	foundPeers := 0
	for {
		select {
		case err := <- chanErr:
			return nil, err
		default:
			n, inboundAddr, err := c.Socket.ReadFromUDP(readBuffer)
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return remConns, fmt.Errorf("%w: timeout after %s with %d peers found: %v", ErrTimeoutDuringServerConnect, c.Timeout.String(), foundPeers, err)
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

			foundPeers = len(remConns)
			if foundPeers == expected {
				return remConns, nil
			}
		}
	}
}

// ConnectPeers should only be used after a preceding Connect has been called and timed out with ErrTimeoutDuringServerConnect.
// This method then allows for trying to connect to remConns (returned by Connect). The method returns ErrTimeoutDuringPeerConnect
// if not all peers respond properly.
//
// This method will refresh the timeout (as it should only be called after Connect has timed out).
// Consider this when configuring the timeout of the server.
func (c client) ConnectPeers(remConns []*net.UDPAddr) error {
	connectionsBuffer := 16

	if c.readDeadline.Before(time.Now()) && c.Timeout >= 0 {
		c.readDeadline = time.Now().Add(c.Timeout)
		if err := c.Socket.SetReadDeadline(c.readDeadline); err != nil {
			return err
		}
	}

	readBuffer := make([]byte, 0xffff)
	remotes := make(map[string]chan string)
	cErr := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer cancel()

	for _, peer := range remConns {
		ch := make(chan string, connectionsBuffer)
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

	msg := syn
	retryPeriod := time.Duration(0)

	send := func() {
		_, err := c.Socket.WriteToUDP([]byte(msg), peer)
		if err != nil {
			cErr <- err
			cancel()
			return
		}}

	for {
		select {
		case <- ctx.Done():
			return
		case str := <- ch:
			if str == syn {
				msg = ack
			} else if str == ack {
				msg = ack
				send()
				go func(delay time.Duration) { // Send once more after a delay to reduce risk of first packet being lost
					time.Sleep(delay)
					send()
				}(c.PeerRetryPeriod)
				wg.Done()
				return
			}
		case <-time.After(retryPeriod):
			retryPeriod = c.PeerRetryPeriod
			send()
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
