package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func connectToServer() (net.Conn, error) {
	for i := 0; i < 3; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err == nil {
			return conn, nil
		}

		fmt.Println("Failed to connect to server, retrying...")
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to server")
}

func main() {
	conn, err := connectToServer()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn.Write([]byte("ping\n"))
		message, err := bufio.NewReader(conn).ReadString('\n')

		fmt.Println("receive message: ", message)

		if err != nil || message != "pong\n" {
			fmt.Println("Lost connection to server, reconnecting...")
			conn, err = connectToServer()
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		time.Sleep(3 * time.Second)
	}
}
