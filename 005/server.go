package main

import (
	"fmt"
	"net"
	"context"
	"io"
	"os"
	"time"
	"os/signal"
	"syscall"
	"bytes"
	"net/http"
	"golang.org/x/sys/unix"
)

const BackendURL = "http://127.0.0.1:50000/hash"
const NumThreads = 64
const ServerPort = 40000
const MaxPacketSize = 1500
const SocketBufferSize = 100*1024*1024

var httpClient *http.Client

func main() {

	fmt.Printf("starting %d server threads on port %d\n", NumThreads, ServerPort)

    httpClient = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 1000000}, Timeout: 1 * time.Second}

	for i := 0; i < NumThreads; i++ {
		go func(threadIndex int) {
			runServerThread(threadIndex)
		}(i)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
}

func runServerThread(threadIndex int) {

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

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:40000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	defer lp.Close()

	if err := conn.SetReadBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket read buffer size: %v", err))
	}

	if err := conn.SetWriteBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket write buffer size: %v", err))
	}

	buffer := make([]byte, MaxPacketSize)

	for {
		packetBytes, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			break
		}
		if packetBytes != 100 {
			continue
		}
		request := buffer[:packetBytes]
		response := PostBinary(BackendURL, request)
		if len(response) != 8 {
			return
		}
		conn.WriteToUDP(response[:], from)
	}	
}

func PostBinary(url string, data []byte) []byte {
	buffer := bytes.NewBuffer(data)
	request, _ := http.NewRequest("POST", url, buffer)
	request.Header.Add("Content-Type", "application/octet-stream")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil
	}
	body, error := io.ReadAll(response.Body)
	if error != nil {
		return nil
	}
	return body
}
