package client

import (
	"fmt"
	"os"
)

var (
	ErrTimeoutDuringServerConnect = fmt.Errorf("%w: timeout during attempting to establish a connection to remote peer to peer mediator server", os.ErrDeadlineExceeded)
)
