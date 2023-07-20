package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Lost connection to client")
			return
		}

		if message == "ping\n" {
			conn.Write([]byte("pong\n"))
		}
	}
}

func main() {
	ln, _ := net.Listen("tcp", ":8080")

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}
