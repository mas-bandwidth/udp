package main

import (
	"fmt"
	"os"
	"io"
	"hash/fnv"
	"net/http"
	"encoding/binary"
)

const BackendPort = 50000
const RequestsPerBlock = 100
const RequestSize = 4 + 2 + 100
const ResponseSize = 4 + 2 + 8
const BlockSize = RequestsPerBlock * RequestSize

func main() {
	fmt.Printf("starting backend on port %d\n", BackendPort)
	http.HandleFunc("/hash", hash)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", BackendPort), nil)
	if err != nil {
		fmt.Printf("error: error starting http server: %v", err)
		os.Exit(1)
	}
}

func hash(w http.ResponseWriter, req *http.Request) {
	request, err := io.ReadAll(req.Body)
	if err != nil || len(request) != BlockSize {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestIndex := 0
	responseIndex := 0
	response := [ResponseSize*RequestsPerBlock]byte{}
	for i := 0; i < RequestsPerBlock; i++ {
		copy(response[responseIndex:responseIndex+6], request[requestIndex:requestIndex+6])
		hash := fnv.New64a()
		hash.Write(request)
		data := hash.Sum64()
		binary.LittleEndian.PutUint64(response[responseIndex+6:responseIndex+6+8], data)
		requestIndex += RequestSize
		responseIndex += ResponseSize
	}
	w.Write(response[:])
}
