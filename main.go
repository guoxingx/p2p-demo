package main

import (
	"flag"
	"fmt"
	"net"
)

var server = flag.String("h", "8000", "host port, default 8000")
var connect = flag.String("c", "0", "connect port")

func main() {
	flag.Parse()
	node := NewNode(*server)
	go node.SendRandMessage()

	if *connect != *server && *connect != "0" {
		raddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%s", *connect))
		fmt.Println(*connect, raddr)
		go node.Connect(raddr)
	}

	node.Run()
}
