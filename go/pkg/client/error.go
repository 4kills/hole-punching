package client

import (
	"fmt"
	"os"
)

var (
	ErrTimeoutDuringServerConnect = fmt.Errorf("%w: timeout during attempting to establish a connection to mediator server", os.ErrDeadlineExceeded)
	ErrTimeoutDuringPeerConnect = fmt.Errorf("%w: timeout during attempting to establish a peer to peer network", os.ErrDeadlineExceeded)
)
