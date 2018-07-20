package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

type Node struct {
	Address   *net.TCPAddr
	Connected map[string]*net.TCPConn
}

type ConnMessage struct {
	Connected bool
	Action    string
	Conn      *net.TCPConn
}

// NewNode creates a new node
func NewNode(port string) *Node {
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%s", port))
	return &Node{addr, map[string]*net.TCPConn{}}
}

// SendRandMessage simulate message sending in tcp connections
func (node *Node) SendRandMessage() {
	for {
		time.Sleep(time.Second)
		randInt := rand.Intn(100)
		if randInt < 20 {
			fmt.Printf("ready to send message...(%v rolled)\n", randInt)
			for _, conn := range node.Connected {
				conn.Write([]byte(fmt.Sprintf("%v from %v", randInt, node.Address)))
			}
		}
	}
}

// Run starts to accept tcp connections.
func (node *Node) Run() {
	listener, err := net.ListenTCP("tcp", node.Address)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()
	fmt.Println("node start at ", node.Address)

	// connection message channel
	connChan := make(chan *ConnMessage)
	// data channel
	dataChan := make(chan []byte)

	go ListenConnections(listener, connChan, dataChan)

	for {
		select {
		case connMessage := <-connChan:
			if connMessage.Connected {
				// new connection, insert into node.Connected
				node.Connected[connMessage.Conn.RemoteAddr().String()] = connMessage.Conn
				fmt.Printf("connect from %v\n", connMessage.Conn.RemoteAddr())
				fmt.Printf("current connected: %v\n", node.Connected)
			} else {
				// lost connect, remove from node.Connected
				delete(node.Connected, connMessage.Conn.RemoteAddr().String())
				fmt.Printf("disconnect from %v\n", connMessage.Conn.RemoteAddr())
				fmt.Printf("current connected: %v\n", node.Connected)
			}
		case data := <-dataChan:
			fmt.Printf("receive data: %s\n", data)
		}
	}
}

func (node *Node) Connect(raddr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		fmt.Printf("unable to connect to %v, [err: %v]\n", raddr, err)
		return
	}
	defer conn.Close()

	// new connection, insert into node.Connected
	node.Connected[conn.RemoteAddr().String()] = conn
	fmt.Printf("connect to %v\n", conn.RemoteAddr())
	fmt.Printf("current connected: %v\n", node.Connected)

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				// lost connect, remove from node.Connected
				delete(node.Connected, conn.RemoteAddr().String())
				fmt.Printf("disconnect to %v\n", conn.RemoteAddr())
				fmt.Printf("current connected: %v\n", node.Connected)
				return
			}
		}
		fmt.Printf("receive data: %s\n", buf[:n])
	}
}

// ListenConnections accepts tcp connections with goroutines
func ListenConnections(listener *net.TCPListener, connChan chan *ConnMessage, dataChan chan []byte) {
	for {
		conn, _ := listener.AcceptTCP()
		go HandleConnection(conn, connChan, dataChan)

		connChan <- &ConnMessage{true, "connect", conn}
	}
}

// HandleConnection accepts data and disconnected.
func HandleConnection(conn *net.TCPConn, connChan chan *ConnMessage, dataChan chan []byte) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				connChan <- &ConnMessage{false, "disconnect", conn}
			} else {
				connChan <- &ConnMessage{false, "unknownError", conn}
			}
			return
		}
		dataChan <- buf[:n]
	}
}
