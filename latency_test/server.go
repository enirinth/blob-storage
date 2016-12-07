package main

import (
    "fmt"
    "log"
    "net"
    "net/rpc"
)

type Handler int

func (p *Handler) Ping(msg string, ack *bool) error {
    fmt.Printf("Ping()\n")
    //fmt.Printf("%s\n", msg)
    *ack = true
    return nil
}

func main() {
    addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:9001")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("[Server Started]\n")
    inbound, err := net.ListenTCP("tcp", addr)
    if err != nil {
        log.Fatal(err)
    }

    handler := new(Handler)
    rpc.Register(handler)
    rpc.Accept(inbound)
}
