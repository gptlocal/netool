package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func main() {
	buf := new(bytes.Buffer)

	buf.WriteByte(0x01)
	buf.WriteByte(0x02)
	buf.Write([]byte{0, 0})

	binary.Write(buf, binary.BigEndian, uint8(0x02))
	binary.Write(buf, binary.BigEndian, uint16(255))

	fmt.Printf("%v\n", buf.Bytes())
	fmt.Printf("%x\n", buf.Bytes())
	printHexWithSpaces(buf.Bytes())
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
