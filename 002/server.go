package main

import (
	"fmt"
	"net"
	"hash/fnv"
	"encoding/binary"
	"os"
	"os/signal"
	"syscall"
)

const ServerPort = 40000
const MaxPacketSize = 1500
const SocketBufferSize = 2*1024*1024

func main() {

	fmt.Printf("starting server on port %d\n", ServerPort)

	addr := net.UDPAddr{
	    Port: ServerPort,
	    IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(fmt.Sprintf("could not create udp socket: %v", err))
	}
	defer conn.Close()

	if err := conn.SetReadBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket read buffer size: %v", err))
	}

	if err := conn.SetWriteBuffer(SocketBufferSize); err != nil {
		panic(fmt.Sprintf("could not set socket write buffer size: %v", err))
	}

	buffer := make([]byte, MaxPacketSize)

	go func() {
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
	}()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
}
