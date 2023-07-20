package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8080")

	for {
		conn.Write([]byte("ping\n"))
		message, _ := bufio.NewReader(conn).ReadString('\n')

		if message == "pong\n" {
			fmt.Println("Received pong from server")
		} else {
			fmt.Println("Lost connection to server")
			return
		}

		time.Sleep(5 * time.Second)
	}
}
