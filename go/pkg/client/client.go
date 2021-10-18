package client

import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "strings"
    "sync"
    "time"
)

const network = "udp"

var timeout = time.Second * 3

func Connect(id []byte, wellKnown string, expected int) ([]*net.UDPConn, error) {
    wellKnownUPD, err := net.ResolveUDPAddr(network, wellKnown)
    if err != nil {
        return nil, err
    }
    wellKnownConn, err := net.DialUDP(network, nil, wellKnownUPD)
    if err != nil {
        return nil, err
    }
    defer wellKnownConn.Close()

    readBuffer := make([]byte, 1 << 16)
    var remConns []*net.UDPConn

    // TODO add max tries/deadline/timeout
    for {
        fmt.Println("Try") // TODO: remove this later
        _, err = io.Copy(wellKnownConn, bytes.NewBuffer(id))
        if err != nil {
            return nil, err
        }

        wellKnownConn.SetReadDeadline(time.Now().Add(timeout))
        n, _, err := wellKnownConn.ReadFromUDP(readBuffer)
        if errors.Is(err, os.ErrDeadlineExceeded) {
            continue
        } else if err != nil {
            return nil, err
        }
        if n == 0 {
            continue
        }

        remConns, err = parse(readBuffer[:n], wellKnownConn.LocalAddr().(*net.UDPAddr))
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
        go connectIndividual(c, wg)
    }

    wg.Wait()

    return remConns, nil
}

func connectIndividual(conn *net.UDPConn, wg *sync.WaitGroup) {
    readBuffer := make([]byte, 1 << 16) // TODO: adjust size

    // TODO: add deadline/ max tries here
    for {
        fmt.Println("Try sending packet to " + conn.RemoteAddr().String()) // TODO: remove this
        // TODO: change "Hello" to something that makes sense
        _, err := io.Copy(conn, bytes.NewBuffer([]byte("Hello")))
        if err != nil {
            panic(err) // TODO: change this behavior to sth sensical
        }

        conn.SetReadDeadline(time.Now().Add(timeout))
        n, _, err := conn.ReadFromUDP(readBuffer)
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

func parse(content []byte, laddr *net.UDPAddr) ([]*net.UDPConn, error) {
    rawAddrs := strings.Split(string(content), ",")
    ret := make([]*net.UDPConn, len(rawAddrs))

    for i, v := range rawAddrs {
        vUDP, err := net.ResolveUDPAddr(network, v)
        if err != nil {
            return ret, err
        }
        if ret[i], err = net.DialUDP(network, laddr, vUDP); err != nil {
            return ret, err
        }
    }

    return ret, nil
}
