package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"sync/atomic"
	"math/rand"
)

const StartPort = 10000
const MaxPacketSize = 1500
const SocketBufferSize = 256*1024*1024

var numClients int

var quit uint64
var packetsSent uint64
var packetsReceived uint64

func GetInt(name string, defaultValue int) int {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(value)
}

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

	serverAddress := GetAddress("SERVER_ADDRESS", "127.0.0.1:40000")

	numClients = GetInt("NUM_CLIENTS", 1)

	fmt.Printf("starting %d clients\n", numClients)

	fmt.Printf("server address is %s\n", serverAddress.String())

	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		go func(clientIndex int) {
			wg.Add(1)
			runClient(clientIndex, &serverAddress)
			wg.Done()
		}(i)
	}

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second)
 
	prev_sent := uint64(0)
	prev_received := uint64(0)

 	for {
		select {
		case <-termChan:
			fmt.Printf("\nreceived shutdown signal\n")
			atomic.StoreUint64(&quit, 1)
	 	case <-ticker.C:
	 		sent := atomic.LoadUint64(&packetsSent)
	 		received := atomic.LoadUint64(&packetsReceived)
	 		sent_delta := sent - prev_sent
	 		received_delta := received - prev_received
	 		fmt.Printf("sent delta %d, received delta %d\n", sent_delta, received_delta)
			prev_sent = sent
			prev_received = received
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
	    IP:   net.ParseIP("0.0.0.0"),
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

	packetData := make([]byte, 100)

	rand.Read(packetData)

	for {
		quit := atomic.LoadUint64(&quit)
		if quit != 0 {
			break
		}
		for i := 0; i < 10; i++ {
			conn.WriteToUDP(packetData[:], serverAddress)
		}
		atomic.AddUint64(&packetsSent, 10)
		time.Sleep(time.Millisecond*100)
	}
}
