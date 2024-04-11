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
	"encoding/binary"
	"golang.org/x/sys/unix"
)

const BackendURL = "http://127.0.0.1:50000/hash"
const NumThreads = 64
const ServerPort = 40000
const MaxPacketSize = 1500
const SocketBufferSize = 100*1024*1024
const RequestsPerBlock = 100
const RequestSize = 4 + 2 + 100
const BlockSize = RequestsPerBlock * RequestSize
const ResponseSize = 4 + 2 + 8

var httpClient *http.Client

type Request struct {
	data []byte
	from net.UDPAddr
}

type RequestGroup struct {
	requests [RequestsPerBlock]Request
}

var channel chan *RequestGroup

var socket [NumThreads]*net.UDPConn

func main() {

	fmt.Printf("starting %d server threads on port %d\n", NumThreads, ServerPort)

    httpClient = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 1000}, Timeout: 1 * time.Second}

    channel = make(chan *RequestGroup)

	for i := 0; i < NumThreads; i++ {
		createServerSocket(i)
	}

	for i := 0; i < NumThreads; i++ {
		go func(threadIndex int) {
			runServerThread(threadIndex)
		}(i)
	}

	go func() {
		runWorkerThread()
	}()

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

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:40000")
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

	buffer := make([]byte, MaxPacketSize)

	requestIndex := 0
	requestGroup := &RequestGroup{}

	for {
		
		packetBytes, from, err := conn.ReadFromUDP(buffer)
		
		if err != nil {
			break
		}
		
		if packetBytes != 100 {
			continue
		}
		
		requestGroup.requests[requestIndex] = Request{data: buffer[:packetBytes], from: *from}

		if requestIndex == RequestsPerBlock {
			channel <- requestGroup
			requestGroup = &RequestGroup{}
			requestIndex = 0
		}
	}	
}

func runWorkerThread() {
	for {
		requestGroup := <- channel
		block := make([]byte, BlockSize)
		index := 0
		for i := 0; i < RequestsPerBlock; i++ {
			request := &requestGroup.requests[i]
			copy(block[index:], request.from.IP.To4())
			binary.LittleEndian.PutUint16(block[index+4:index+6], uint16(request.from.Port))
			copy(block[index+6:index+RequestSize], request.data)
			index += RequestSize
		}
		go func() {
			response := PostBinary(BackendURL, block)
			if len(response) == ResponseSize * RequestsPerBlock {
				responseIndex := 0
				for i := 0; i < RequestsPerBlock; i++ {
					ip := response[responseIndex:responseIndex+4]
					port := binary.LittleEndian.Uint16(response[responseIndex+4:responseIndex+6])
					from := net.UDPAddr{IP: ip, Port: int(port)}
					socketIndex := i % NumThreads
					socket[socketIndex].WriteToUDP(response[responseIndex+6:responseIndex+6+8], &from)
					responseIndex += ResponseSize
				}
			}
		}()
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