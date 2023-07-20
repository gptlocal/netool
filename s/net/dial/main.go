package main

import (
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: www.baidu.com\r\n\r\n")); err != nil {
		log.Fatalf("err: %v", err)
	}

	recv := make([]byte, 4096)
	if _, err := conn.Read(recv); err != nil {
		log.Fatalf("err: %v", err)
	}

	log.Printf("recv:\n%s", recv)
}
