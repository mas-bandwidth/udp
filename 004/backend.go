package main

import (
	"fmt"
	// "github.com/gorilla/mux"
)

const BackendPort = 50000

func main() {
	fmt.Printf("starting backend on port %d\n", BackendPort)
	// 
}

/*
		hash := fnv.New64a()
		hash.Write(buffer[:packetBytes])
		data := hash.Sum64()
		responsePacket := [8]byte{}
		binary.LittleEndian.PutUint64(responsePacket[:], data)
*/
