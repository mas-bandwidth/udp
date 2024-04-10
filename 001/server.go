package main

import (
	"fmt"
	"net"
	"hash/fnv"
	"encoding/binary"
)

const ServerPort = 40000
const MaxPacketSize = 1500

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

	buffer := make([]byte, MaxPacketSize)

	for {
		packetBytes, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("udp receive error: %v\n", err)
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
