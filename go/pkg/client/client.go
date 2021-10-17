package client

import (
    "bytes"
    "errors"
    "io"
    "net"
    "os"
    "time"
)

const wellKnown = "127.0.0.1:5000"

const network = "udp"

var timeout = time.Second * 3

func connect(id []byte) error {
    wellKnownUPD, err := net.ResolveUDPAddr(network, wellKnown)
    if err != nil {
        return err
    }
    wellKnownConn, err := net.DialUDP(network, nil, wellKnownUPD)
    if err != nil {
        return err
    }
    defer wellKnownConn.Close()

    readBuffer := make([]byte, 1 << 16)

    // TODO add max tries/deadline/timeout
    for {
        _, err = io.Copy(wellKnownConn, bytes.NewBuffer(id))
        if err != nil {
            return err
        }

        wellKnownConn.SetReadDeadline(time.Now().Add(timeout))
        n, _, err := wellKnownConn.ReadFromUDP(readBuffer)
        if errors.Is(err, os.ErrDeadlineExceeded) {
            continue
        } else if err != nil {
            return err
        }
    }
}

func parse()
