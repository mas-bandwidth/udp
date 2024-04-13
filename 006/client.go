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

const NumClients = 20000
const MaxPacketSize = 1500

var quit uint64
var packetsSent uint64
var packetsReceived uint64

func ParseAddress(input string) net.UDPAddr {
	address := net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func main() {

	fmt.Printf("starting %d clients\n", NumClients)

	serverAddress := ParseAddress("127.0.0.1:40000")

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
	    Port: 0,
	    IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return
	}
	defer conn.Close()

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
