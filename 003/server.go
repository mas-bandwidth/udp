package main

import (
	"fmt"
	"net"
	"hash/fnv"
	"encoding/binary"
	"context"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/sys/unix"
)

const NumThreads = 64
const ServerPort = 40000
const MaxPacketSize = 1500
const SocketBufferSize = 100*1024*1024

func main() {

	fmt.Printf("starting %d server threads on port %d\n", NumThreads, ServerPort)

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
		hash := fnv.New64a()
		hash.Write(buffer[:packetBytes])
		data := hash.Sum64()
		responsePacket := [8]byte{}
		binary.LittleEndian.PutUint64(responsePacket[:], data)
		conn.WriteToUDP(responsePacket[:], from)
	}	
}
