package main

import (
	"encoding/binary"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
	bytes, err := proto.Marshal(wrapperspb.Int32(65537))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", int32(65537))
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	bytes, err = proto.Marshal(wrapperspb.UInt32(65537))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", uint32(65537))
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	bytes, err = proto.Marshal(wrapperspb.Int64(65537))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", int64(65537))
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	bytes, err = proto.Marshal(wrapperspb.UInt64(65537))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", uint64(65537))
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	int32bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(int32bytes, 31)
	bytes, err = proto.Marshal(wrapperspb.Bytes(int32bytes))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", int32bytes)
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	bytes, err = proto.Marshal(wrapperspb.Bool(false))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", false)
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	bytes, err = proto.Marshal(wrapperspb.Bool(true))
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%x\n", bytes)
	fmt.Printf("%v: ", true)
	for _, b := range bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
}
