//go:build linux

package main

import (
	"fmt"
	"syscall"
)

func main() {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println("Error in syscall.Socket:", err)
		return
	}

	defer syscall.Close(fd)

	err = syscall.SetNonblock(fd, true)
	if err != nil {
		fmt.Println("Error in syscall.SetNonblock:", err)
		return
	}

	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		fmt.Println("Error in syscall.EpollCreate1:", err)
		return
	}

	defer syscall.Close(epfd)

	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN,
	}

	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, event)
	if err != nil {
		fmt.Println("Error in syscall.EpollCtl:", err)
		return
	}

	events := make([]syscall.EpollEvent, 10)
	for {
		nevents, err := syscall.EpollWait(epfd, events, -1)
		if err != nil {
			fmt.Println("Error in syscall.EpollWait:", err)
			return
		}

		for ev := 0; ev < nevents; ev++ {
			if events[ev].Fd == int32(fd) {
				fmt.Println("Received an event on fd")
			}
		}
	}
}
