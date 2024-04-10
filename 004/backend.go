package main

import (
	"fmt"
	"net/http"
)

const BackendPort = 50000

func main() {
	fmt.Printf("starting backend on port %d\n", BackendPort)
	http.HandleFunc("/hash", hash)
	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", BackendPort), nil)
	if err != nil {
		core.Error("error starting http server: %v", err)
		os.Exit(1)
	}
}

func hash(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

/*
		hash := fnv.New64a()
		hash.Write(buffer[:packetBytes])
		data := hash.Sum64()
		responsePacket := [8]byte{}
		binary.LittleEndian.PutUint64(responsePacket[:], data)
*/
