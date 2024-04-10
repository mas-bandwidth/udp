package main

import (
	"fmt"
	"os"
	"io"
	"net/http"
)

const BackendPort = 50000

func main() {
	fmt.Printf("starting backend on port %d\n", BackendPort)
	http.HandleFunc("/hash", hash)
	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", BackendPort), nil)
	if err != nil {
		fmt.Printf("error: error starting http server: %v", err)
		os.Exit(1)
	}
}

func hash(w http.ResponseWriter, req *http.Request) {
	request, err := io.ReadAll(req.Body)
	if err != nil || len(request) != 100 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash := fnv.New64a()
	hash.Write(request)
	data := hash.Sum64()
	response := [8]byte{}
	binary.LittleEndian.PutUint64(responsePacket[:], data)
	fmt.Fwrite(w, response)
	w.WriteHeader(http.StatusOK)
	return
}
