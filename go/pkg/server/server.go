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

	log logr.Logger
  	conn *net.UDPConn
}

// New constructs a default server listening to listeningAddr with a Go map as server.AddrStore implementation and
// the Go standard log package as logr.Logger.
// It is strongly recommended to review the server.DomainTimeout field.
func New(listeningAddr string) (server, error) {
	s := server{
		ListeningAddr: listeningAddr,
		DomainTimeout: 1 * time.Hour,
		MaxPacketSize: 1024,
		AddrStore:     domainAddrMap{make(map[string][]string), &sync.Mutex{}},

		log: stdr.New(nil),
	}

	addr, err := net.ResolveUDPAddr(udpNetworkName, listeningAddr)
	if err != nil {
		return s, err
	}

	s.conn, err = net.ListenUDP(udpNetworkName, addr)
	if err != nil {
		return s, err
	}

	return s, err
}

// ListenAndServe starts the server, listening to server.ListeningAddr and handling inbound packets.
func (s server) ListenAndServe() {
	buffer := make([]byte, 2 * s.MaxPacketSize)

	for {
		n, addr, err := s.conn.ReadFromUDP(buffer)
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
	_, err = s.conn.WriteToUDP([]byte(payload), addr)
	if err != nil {
		s.log.Error(err, "writing to remote address ; socket listening on port", logKeyAddr, addr.String(), "port", s.conn.RemoteAddr().String())
		return
	}

	s.log.V(1).Info("wrote package to address with payload", logKeyAddr, addr.String(), "payload", payload)
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

func (s server) GetLogger() logr.Logger {
	return s.log
}