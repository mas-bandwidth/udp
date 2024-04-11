package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"os"
	"os/signal"
	"syscall"
	"sync/atomic"
	"math/rand"
)

const StartPort = 10000
const NumClients = 30000
const MaxPacketSize = 1500
const SocketBufferSize = 100*1024*1024

var quit uint64
var packetsSent uint64
var packetsReceived uint64

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

	fmt.Printf("starting %d clients\n", NumClients)

	serverAddress := GetAddress("SERVER_ADDRESS", "127.0.0.1:40000")

	var wg sync.WaitGroup

	for i := 0; i < NumClients; i++ {
		go func(clientIndex int) {
			wg.Add(1)
			runClient(clientIndex, &serverAddress)
			wg.Done()
		}(i)
	}

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second)
 
 	for {
		select {
		case <-termChan:
			fmt.Printf("\nreceived shutdown signal\n")
			atomic.StoreUint64(&quit, 1)
	 	case <-ticker.C:
	 		sent := atomic.LoadUint64(&packetsSent)
	 		received := atomic.LoadUint64(&packetsReceived)
	 		fmt.Printf("sent %d, received %d (%.1f%%)\n", sent, received, float64(received)/float64(sent)*100.0)
	 	}
		quit := atomic.LoadUint64(&quit)
		if quit != 0 {
			break
		}
 	}

	fmt.Printf("shutting down\n")

	wg.Wait()	

	fmt.Printf("done.\n")
}

func runClient(clientIndex int, serverAddress *net.UDPAddr) {

	addr := net.UDPAddr{
	    Port: StartPort + clientIndex,
	    IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return // IMPORTANT: to get as many clients as possible on one machine, if we can't bind to a specific port, just ignore and carry on
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
			packetBytes, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				break
			}
			if packetBytes != 8 {
				continue
			}
			atomic.AddUint64(&packetsReceived, 1)
		}
	}()

	for {
		quit := atomic.LoadUint64(&quit)
		if quit != 0 {
			break
		}
		packetData := make([]byte, 100)
		rand.Read(packetData)
		conn.WriteToUDP(packetData[:], serverAddress)
		atomic.AddUint64(&packetsSent, 1)
		time.Sleep(time.Millisecond*10)
	}
}
