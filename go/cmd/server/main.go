package main

import (
    "github.com/4kills/hole-punching/go/pkg/server"
    "github.com/go-logr/stdr"
    "os"
)

func main() {
    wellKnownAddr := ":5000"
    if len(os.Args) > 1 {
        wellKnownAddr = os.Args[1]
    }

    s, err := server.New(wellKnownAddr)
    if err != nil {
        panic(err)
    }

    stdr.SetVerbosity(1)

    s.ListenAndServe()
}