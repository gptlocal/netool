package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func main() {
	resolver := &net.Resolver{
		PreferGo: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()

	ips, err := resolver.LookupIPAddr(ctx, "www.google.com")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, ip := range ips {
		fmt.Println(ip.String())
	}
}
