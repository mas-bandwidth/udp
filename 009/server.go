package main

import (
	"fmt"
	"net"
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"bytes"
	"net/http"
	"golang.org/x/sys/unix"
)

const NumThreads = 64
const ServerPort = 40000
const SocketBufferSize = 1024*1024*1024

var socket [NumThreads]*net.UDPConn

func main() {

	fmt.Printf("starting %d server threads on port %d\n", NumThreads, ServerPort)

	for i := 0; i < NumThreads; i++ {
		createServerSocket(i)
	}

	for i := 0; i < NumThreads; i++ {
		go func(threadIndex int) {
			runServerThread(threadIndex)
		}(i)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
}

func createServerSocket(threadIndex int) {

	lc := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			err := c.Control(func(fileDescriptor uintptr) {
				err := unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse port socket option: %v", err))
				}
			})
			return err
		},
	}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:40000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	socket[threadIndex] = conn
}

func runServerThread(threadIndex int) {

	conn := socket[threadIndex]

	defer conn.Close()

	if err := conn.SetReadBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket read buffer size: %v", err))
	}

	if err := conn.SetWriteBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket write buffer size: %v", err))
	}

	buffer := make([]byte, 1500)

	for {

		packetBytes, from, err := conn.ReadFromUDP(buffer[:])
		if err != nil {
			break
		}
		
		if packetBytes != 100 {
			continue
		}

		var dummy [8]byte 
		socket[threadIndex].WriteToUDP(dummy[:], from)
	}	
}

func PostBinary(client *http.Client, url string, data []byte) []byte {
	buffer := bytes.NewBuffer(data)
	request, _ := http.NewRequest("POST", url, buffer)
	request.Header.Add("Content-Type", "application/octet-stream")
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("error: posting request: %v\n", err)
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		fmt.Printf("error: status code is %d\n", response.StatusCode)
		return nil
	}
	body, error := io.ReadAll(response.Body)
	if error != nil {
		fmt.Printf("error: reading response: %v\n", error)
		return nil
	}
	return body
}
