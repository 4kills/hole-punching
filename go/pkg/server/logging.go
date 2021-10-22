package server

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"io"
	stdlog "log"
)

const(
	logKeyAddr = "address"
)

var log = stdr.New(nil)

// SetLogger takes a logr.Logger. If logger is nil or not of type logr.Logger, logs will be discarded and not put anywhere.
// The default logger of this library uses the default Go log implementation and writes to std streams.
func SetLogger(logger interface{}) {
	l, ok := logger.(logr.Logger)
	if !ok {
		log = stdr.New(stdlog.New(io.Discard, "", 0))
		return
	}

	log = l
}
