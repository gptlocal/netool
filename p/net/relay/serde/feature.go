package main

import (
	"bytes"
	"fmt"
	"github.com/gptlocal/netool/p/net/relay"
	"github.com/gptlocal/netool/p/net/relay/features"
)

func main() {
	req := relay.Request{
		Version: relay.Version1,
		Cmd:     relay.CmdConnect,
		Features: []features.Feature{
			&features.AddrFeature{
				AType: features.AddrIPv4,
				Host:  "127.0.0.1",
				Port:  8080,
			},
		},
	}
	fmt.Printf("%v, %v\n", req, req.Features[0])

	w := new(bytes.Buffer)
	req.WriteTo(w)
	fmt.Printf("%v\n", w.Bytes())
	printHexWithSpaces(w.Bytes())
	fmt.Printf("%x\n", w.Bytes())

	fb, _ := req.Features[0].Encode()
	fmt.Printf("%v\n", fb)
	printHexWithSpaces(fb)

	fmt.Println("==================")

	req2 := &relay.Request{}
	req2.ReadFrom(bytes.NewReader(w.Bytes()))
	fmt.Printf("%v, %v\n", req2, req2.Features[0])
}

func printHexWithSpaces(p []byte) {
	for i, b := range p {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%02x", b)
	}
	fmt.Println()
}
