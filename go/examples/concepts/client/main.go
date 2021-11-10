package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "time"
)

const network = "udp"
const serverHost = "dominik-ochs.de:5000"

func main() {
    conn, peer := establishConnection()

    go func() {
        b := make([]byte, 0xffff)
        n, _, _ := conn.ReadFromUDP(b)
        fmt.Println("Peer:", b[:n])
    }()

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        msg := scanner.Text()
        conn.WriteToUDP([]byte(msg), peer)
        fmt.Println("Me:", msg)
    }
}

func establishConnection() (*net.UDPConn, *net.UDPAddr) {
    conn, _ := net.ListenUDP(network, nil)

    serverAddr, _ := net.ResolveUDPAddr(network, serverHost)
    // Write to server
    conn.WriteToUDP([]byte{}, serverAddr)

    // Listen for server response with peer's remote address
    b := make([]byte, 128)
    n, _, _ := conn.ReadFromUDP(b)

    // fetch remote addr
    peerAddr, _ := net.ResolveUDPAddr(network, string(b[:n]))

    // TODO: does this work?
    // send datagram to peer
    for i := 0; i < 2; i++ {
        time.Sleep(time.Millisecond * 70)
        conn.WriteToUDP([]byte{}, peerAddr)
    }

    return conn, peerAddr
}