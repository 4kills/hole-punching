# Hole Punching

Hole punching is the technique of establishing connections between computers
from different local networks over the Internet. For further reading on hole punching see [Wikipedia.](https://en.wikipedia.org/wiki/Hole_punching_(networking))

The goal of this project is to provide libraries for implementing peer-to-peer clients and the rendezvous server in several languages.
These libraries will perform the hole punching and establish the connections.
After that you may do as you please with these connections.

Feel free to navigate to the respective language directories for more information 
as well as to the [docs](./docs) for an overview and the [basic concepts of hole punching](./docs/README.md). 

## Supported Protocols and Languages
- UDP
  - Server
    - [x] Go
  - Client
    - [x] Go
    - [ ] C#
- TCP
  - Server
    - [ ] Go
  - Client
    - [ ] Go
    - [ ] C#

## UDP Hole Punching Libraries

The libraries are designed to all work in the same way, so each server-client combination
(e.g. C# client and Go server) will be valid and work together.  

All implementations will follow this pattern:

| Step | Client   |  Server      |
|----|---|-------|
| 0. | The client will be given the server address, an ID to identify the connection session and a number of expected clients to join | |
| 1. | A client will register to the server by sending a UDP datagram containing the ID used to identify the session. | 
| 2. | | The server will store the client's endpoint and respond with a list of endpoints that have registered using ID |
| 3. | | The server will also notify each endpoint in that list of the new registered endpoint |
| 4. | The client will wait for the expected number of endpoints. When this number is reached the clients will attempt to connect to each other. When the number is not reached, this phase may optionally time out. The user can then decide to proceed with peer connection anyway. | |

5. From this point onward, the server will not be involved any longer and the clients will start sending UDP packets 
to each other with a TCP-like scheme to ensure holes have been punched.
6. When each client has been connected to each other in the endpoint list and connection has been approved, the socket and connections may be used by the consumer.


## Origin

This project originated as a graded project for software engineering class 
and aimed to introduce my classmates to UDP hole punching. 