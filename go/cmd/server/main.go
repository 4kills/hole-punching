package main

import (
    "github.com/4kills/hole-punching/pkg/server"
    "github.com/go-logr/stdr"
)

func main() {
    s, err := server.New(":5000")
    if err != nil {
        panic(err)
    }

    stdr.SetVerbosity(1)

    s.ListenAndServe()
}