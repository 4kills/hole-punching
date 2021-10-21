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

var timeout = time.Second * 3

func Connect(id []byte, wellKnown string, expected int) ([]*net.UDPAddr, error) {
    wellKnownUPD, err := net.ResolveUDPAddr(network, wellKnown)
    if err != nil {
        return nil, err
    }

    socket, err := net.ListenUDP(network, &net.UDPAddr{}) // TODO: change this to listenUDP
    if err != nil {
        return nil, err
    }
    defer socket.Close()

    readBuffer := make([]byte, 1 << 16)
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
        if n == 0 {
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
    readBuffer := make([]byte, 1 << 16) // TODO: adjust size

    // TODO: add deadline/ max tries here
    for {
        fmt.Println("Try sending packet to " + conn.String()) // TODO: remove this
        // TODO: change "Hello" to something that makes sense
        _, err := socket.WriteToUDP([]byte("Hello"), conn)
        if err != nil {
            panic(err) // TODO: change this behavior to sth sensical
        }

        // TODO: deadline wont work across multiple goroutines simultaneously
        socket.SetReadDeadline(time.Now().Add(timeout))
        n, _, err := socket.ReadFromUDP(readBuffer)
        log.Println("listening for packet: error:", err )
        if errors.Is(err, os.ErrDeadlineExceeded) {
            continue
        } else if err != nil {
            panic(err)
        }

        // SUCCESS, connection established
        readBuffer = readBuffer[:n]
        fmt.Println(string(readBuffer))
        break
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
