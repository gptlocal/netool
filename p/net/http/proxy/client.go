//go:build client

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// gost -L=http://:3128
// curl -x http://192.168.1.110:3128/ https://www.google.com

func main() {
	proxy, err := url.Parse("http://192.168.1.110:3128")
	if err != nil {
		panic(err)
	}

	args := os.Args
	if len(args) < 2 {
		panic("usage: http_proxy url")
	}

	targetUrl := os.Args[1]
	cli := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := cli.Get(targetUrl)
	fmt.Println(resp, err)
}
