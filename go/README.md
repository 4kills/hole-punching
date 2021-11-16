# Go Implementation

In this sub-directory, the [libraries for hole punching (client and server) in Go](./pkg),
as well as an [example peer-to-peer chat](./cmd) using those libraries can be found.  
Additionally, the [examples directory](./examples) contains further examples as well as a very basic 
implementation of the [principles](./examples/concepts) that make UDP hole punching work. 

# The Libraries

These libraries are designed as described [here](../README.md#udp-hole-punching-libraries).
This document will add the language specific details. 

## Client 

To create a new configurable client instance with a given rendezvous server: 
```go 
c, _ := client.New("well-known.rendezvous.com:5000")
```

The returned client can then be customized by changing its attributes. 
However, sensible defaults are already set. All the attributes are well-documented in Godoc.  
```go
c.Timeout = time.Second * 30
c.PeerRetryPeriod = time.Millisecond * 20
```

After that, connection may be established with the ID as well as excpected number of peers.
If it doesn't time out, addrs will contain all the endpoints and socket must be used to talk to these endpoints.
```go
addrs, socket, _ := c.Connect([]byte(id), numPeers)
```

Other public functions and methods are well documented with Godoc and should be fairly easy and straightforward to use.

## Server

To create a new server that will listen to port 5000 do the following: 

```go 
s, _ := server.New(":5000")
```

`s` can now be configured in the same fashion as the [client](#client). 
Also consider the logging solution. The default will only log severe errors via std error. 
This behavior can be changed with a [logr](https://github.com/go-logr/logr) implementation.

The server can then be started like this:
```go
s.ListenAndServe()
```

As with the clients, all functions are well documented in code via Godoc. 

# Peer-to-Peer Chat Example

The p2p chat is a very basic example of the usage of this library. It is rudimentary and not encrypted.
It was designed solely for demonstration purposes. 

## Chat Client

The [chat client](./cmd/client) takes three mandatory arguments:
```shell
client <rendezvous> <domain_id> <num_peers>
# or 
go run ./cmd/client <rendezvous> <domain_id> <num_peers>
```

with

| Parameter | Description | Domain |
|:---------:|:-----------:|:------:|
| rendezvous | well-known address of the rendezvous server | <code>(\<IP>&#124;\<FQDN>)?:\<port></code>  |  
| domain_id | id which is used by the clients to identify their chat session | `.+` |
| num_peers | number of participants expected to join the chat session | positive integer | 


## Rendezvous Server

The [server](./cmd/server) can be run directly using Go or using the [docker-compose](./deployments/docker-compose.dev.yml).

Either way, the first command line argument to the program optionally sets the listening address (laddr). 
It defaults to `:5000`.  
Instead of changing the docker-compose and Dockerfile, simply change the port mapping if you require another port.  
