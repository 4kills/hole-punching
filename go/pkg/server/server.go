package server

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type server struct {
	// ListeningAddr is the addr (ip:port) this server listens to.
	ListeningAddr string
	// AddrStore temporarily stores the connecting addresses with the given domain. This can be overridden by your own implementation.
	// E.g. to make it work in a load balanced environment.
	AddrStore AddressStore
	// DomainTimeout is the time after which an address associated to a domain is removed from the server.AddrStore.
	// If DomainTimeout is negative, no addresses are removed.
	DomainTimeout time.Duration
	// MaxPacketSize defines the max length of the packet payload. As the UUID of the client (addr) is 16 bytes, MaxPacketSize - 16 bytes are left for the domain ID.
	// If a packet's payload length exceeds MaxPacketSize, the packet is dropped and not processed.
	MaxPacketSize int

	keepAlive time.Duration
	log    logr.Logger
  	socket *net.UDPConn
}

// New constructs a default server listening to listeningAddr with a Go map as server.AddrStore implementation and
// the Go standard log package as logr.Logger.
// It is strongly recommended reviewing the server.DomainTimeout field.
func New(listeningAddr string) (server, error) {
	s := server{
		ListeningAddr: listeningAddr,
		DomainTimeout: 40 * time.Second,
		keepAlive: 10 * time.Second,
		MaxPacketSize: 1024,
		AddrStore:     domainAddrMap{make(map[string][]string), &sync.Mutex{}, make([]string, 1024)},

		log: stdr.New(nil),
	}

	addr, err := net.ResolveUDPAddr(udpNetworkName, listeningAddr)
	if err != nil {
		return s, err
	}

	s.socket, err = net.ListenUDP(udpNetworkName, addr)
	if err != nil {
		return s, err
	}

	return s, err
}

// ListenAndServe starts the server, listening to server.ListeningAddr and handling inbound packets. Once this is called,
// changes on s are not guaranteed to have an effect.
func (s server) ListenAndServe() {
	s.log.V(1).Info("server started")
	buffer := make([]byte, 2 * s.MaxPacketSize)

	go s.sendKeepAlives()

	for {
		n, addr, err := s.socket.ReadFromUDP(buffer)
		if err != nil {
			s.log.Error(err, "read from udp with remote address: rejecting address", logKeyAddr, addr.String())
			continue
		}
		if n > s.MaxPacketSize {
			s.log.V(1).Info( "package payload by remote address with messageLength bytes exceeded maxPacketSize: rejecting address.",
				logKeyAddr, addr.String(), "messageLength", n, "maxPacketSize", s.MaxPacketSize)
			continue
		}

		id := string(buffer[:n])

		go s.handleConnection(id, addr)
	}
}

func (s server) handleConnection(id string, addr *net.UDPAddr) {
	remoteAddrs, err := s.AddrStore.ProcessAddress(id, addr.String(), s.DomainTimeout)
	if err != nil {
		s.log.Error(err, "could not store address: rejecting address", logKeyAddr, addr.String())
		return
	}

	payload := strings.Join(remoteAddrs, ",")
	_, err = s.socket.WriteToUDP([]byte(payload), addr)
	if err != nil {
		s.log.Error(err, "writing to remote address ; socket listening on port", logKeyAddr, addr.String(), "port", s.socket.RemoteAddr().String())
		return
	}

	s.log.V(1).Info("wrote package to address with payload", logKeyAddr, addr.String(), "payload", payload)
}

func (s server) sendKeepAlives() {
	if s.keepAlive < 0 {
		return
	}

	// TODO: optimize this to not send all packets at once
	for {
		time.Sleep(s.keepAlive)
		s.log.V(1).Info("sending keep-alive packets")

		addrs, err := s.AddrStore.FetchAllAddresses()
		if err != nil {
			s.log.Error(err, "could not fetch addresses")
			continue
		}

		for _, addrStr := range addrs {
			addr, err := net.ResolveUDPAddr(udpNetworkName, addrStr)
			if err != nil {
				s.log.V(1).Error(err, "could not resolve address when trying to send keep alive packet", logKeyAddr, addrStr)
				continue
			}

			_, err = s.socket.WriteToUDP([]byte{}, addr)
			if err != nil {
				s.log.Error(err, "could not write to udp while trying to send keep alive packet, skipping for now", logKeyAddr, addr)
				continue
			}
		}
	}
}

// SetKeepAlive sets the time after which an address receives a keep alive packet in order to keep the NAT mapping intact.
// If the value is negative, keep alive packets are disabled. t must not be greater than or equal 0 but be less than 1 s. If it is, it will be set to 1 s.
func (s server) SetKeepAlive(t time.Duration) {
	if 0 <= t && t < time.Second {
		t = time.Second
	}
	s.keepAlive = t
}

func (s server) KeepAlive() time.Duration {
	return s.keepAlive
}

// SetLogger takes a logr.Logger. If logger is nil or not of type logr.Logger, logs will be discarded and not put anywhere.
// The default logger of this library uses the default Go log implementation and writes to std streams.
func (s server) SetLogger(logger interface{}) {
	l, ok := logger.(logr.Logger)
	if !ok {
		s.log = stdr.New(log.New(io.Discard, "", 0))
		return
	}

	s.log = l
}

func (s server) Logger() logr.Logger {
	return s.log
}