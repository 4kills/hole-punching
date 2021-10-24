package main

import "github.com/4kills/hole-punching/pkg/server"

func main() {
    s, err := server.New(":5000")
    if err != nil {
        panic(err)
    }

    s.ListenAndServe()
}