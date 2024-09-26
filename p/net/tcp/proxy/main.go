package main

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
)

var (
	tinyBufferSize   = 512
	smallBufferSize  = 2 * 1024  // 2KB small buffer
	mediumBufferSize = 8 * 1024  // 8KB medium buffer
	largeBufferSize  = 32 * 1024 // 32KB large buffer
)

var (
	sPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, smallBufferSize)
		},
	}
	mPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, mediumBufferSize)
		},
	}
	lPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, largeBufferSize)
		},
	}
)

func main() {
	args := os.Args
	if len(args) < 2 {
		panic("usage: main url")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 这里指定本地监听地址和远程服务器地址
	go startProxy("localhost:8080", "remote.server.com:80")

	wg.Wait() // 防止main函数立即退出
}

func startProxy(localAddr string, remoteAddr string) {
	// 监听本地地址
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", localAddr, err)
	}
	log.Printf("Listening on %s, proxying to %s", localAddr, remoteAddr)

	for {
		// 接受来自客户端的连接
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// 连接到远程服务器
		serverConn, err := net.Dial("tcp", remoteAddr)
		if err != nil {
			log.Printf("Failed to connect to server %s: %v", remoteAddr, err)
			clientConn.Close()
			continue
		}

		// 启动双向传输
		go func() {
			defer clientConn.Close()
			defer serverConn.Close()
			if err := transport(clientConn, serverConn); err != nil {
				log.Printf("Error in transport: %v", err)
			}
		}()
	}
}

func transport(rw1, rw2 io.ReadWriter) error {
	errc := make(chan error, 1)
	go func() {
		errc <- copyBuffer(rw1, rw2)
	}()

	go func() {
		errc <- copyBuffer(rw2, rw1)
	}()

	if err := <-errc; err != nil && err != io.EOF {
		return err
	}

	return nil
}

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := lPool.Get().([]byte)
	defer lPool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
