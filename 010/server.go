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

const NumThreads = 32
const ServerPort = 40000
const MaxPacketSize = 1500
const SocketBufferSize = 1024*1024*1024
const RequestsPerBlock = 100
const RequestSize = 4 + 2 + 100
const BlockSize = RequestsPerBlock * RequestSize
const ResponseSize = 4 + 2 + 8

type Request struct {
	data [100]byte
	from net.UDPAddr
}

type RequestGroup struct {
	requests [RequestsPerBlock]Request
}

var socket [NumThreads]*net.UDPConn

var channel [NumThreads]chan *RequestGroup

var backendAddress net.UDPAddr

func GetAddress(name string, defaultValue string) net.UDPAddr {
	valueString, ok := os.LookupEnv(name)
	if !ok {
	    valueString = defaultValue
	}
	value, err := net.ResolveUDPAddr("udp", valueString)
	if err != nil {
		panic(fmt.Sprintf("invalid address in envvar %s", name))
	}
	return *value
}

func main() {

	fmt.Printf("starting %d server threads on port %d\n", NumThreads, ServerPort)

	backendAddress = GetAddress("BACKEND_ADDRESS", "127.0.0.1:50000")

	fmt.Printf("backend address is %s\n", backendAddress.String())

	for i := 0; i < NumThreads; i++ {
	    channel[i] = make(chan *RequestGroup)
	}

	for i := 0; i < NumThreads; i++ {
		createServerSocket(i)
	}

	for i := 0; i < NumThreads; i++ {
		go func(threadIndex int) {
			runServerThread(threadIndex)
		}(i)
	}

	for i := 0; i < NumThreads; i++ {
		go func(threadIndex int) {
			runWorkerThread(threadIndex)
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

	requestIndex := 0
	requestGroup := &RequestGroup{}

	for {

		packetBytes, from, err := conn.ReadFromUDP(requestGroup.requests[requestIndex].data[:])
		if err != nil {
			break
		}
		
		if packetBytes != 100 {
			continue
		}

		requestGroup.requests[requestIndex].from = *from
		
		requestIndex++

		if requestIndex == RequestsPerBlock {
			channel[threadIndex] <- requestGroup
			requestGroup = &RequestGroup{}
			requestIndex = 0
		}
	}	
}

func runWorkerThread(threadIndex int) {
	backendURL := fmt.Sprintf("http://%s/hash", backendAddress.String())
	fmt.Printf("backend url is %s\n", backendURL)
    httpClient := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 1000}, Timeout: 1 * time.Second}
	for {
		requestGroup := <- channel[threadIndex]
		block := make([]byte, BlockSize)
		index := 0
		for i := 0; i < RequestsPerBlock; i++ {
			request := &requestGroup.requests[i]
			copy(block[index:], request.from.IP.To4())
			binary.LittleEndian.PutUint16(block[index+4:index+6], uint16(request.from.Port))
			copy(block[index+6:index+RequestSize], request.data[:])
			index += RequestSize
		}
		go func() {
			response := PostBinary(httpClient, backendURL, block)
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

func PostBinary(client *http.Client, url string, data []byte) []byte {
	buffer := bytes.NewBuffer(data)
	request, _ := http.NewRequest("POST", url, buffer)
	request.Header.Add("Content-Type", "application/octet-stream")
	response, err := client.Do(request)
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
