# Go Implementation

In this sub-directory, the [libraries for hole punching (client and server) in Go](./pkg),
as well as an [example peer-to-peer chat](./cmd) using those libraries can be found.  
Additionally, the [examples directory](./examples) contains further examples as well as a very basic 
implementation of the [principles](./examples/concepts) that make UDP hole punching work. 

# The Libraries



# Peer-to-Peer Chat Example

## Chat Client

The [chat client](./cmd/client) takes three mandatory arguments:
```shell
client <rendezvous> <domain_id> <num_peers>
# or 
go run . <rendezvous> <domain_id> <num_peers>
```

with

| Parameter | Description | Domain |
|:---------:|:-----------:|:------:|
| rendezvous | well-known address of the rendezvous server | `(<IP>|<FQDN>)?:<port>`  |  
| domain_id | id which is used by the clients to identify their chat session | `.+` |
| num_peers | number of participants expected to join the chat session | positive integer | 


## Rendezvous Server

The [server](./cmd/server) can be run directly using Go or using the [docker-compose](./deployments/docker-compose.dev.yml).

Either way, the first command line argument to the program optionally sets the listening address (laddr). 
It defaults to `:5000`.  
Instead of changing the docker-compose and Dockerfile, simply change the port mapping if you require another port.  
