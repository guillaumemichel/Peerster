package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	u "github.com/guillaumemichel/Peerster/utils"
)

// RequireMessage : requires a message
func RequireMessage(msg *string) {
	if *msg == "" {
		fmt.Println("Error: message required!")
		os.Exit(1)
	}
}

func main() {
	// treating the flags
	UIPort := flag.String("UIPort", "8080", "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	destination := "127.0.0.1"
	RequireMessage(msg)

	// parse destination
	ip := net.ParseIP(destination)

	// parse port
	port, err := strconv.Atoi(*UIPort)
	if err != nil || port < 0 || port > 65535 {
		log.Fatalf("Error: invalid port %s", *UIPort)
	}

	// creating destination address
	address := net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	// creating upd connection
	udpConn, err := net.DialUDP("udp4", nil, &address)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	packetToSend := u.Message{Text: *msg}

	// serializing the packet to send
	bytesToSend := u.ProtobufMessage(&packetToSend)

	// sending the packet over udp
	_, err = udpConn.Write(bytesToSend)
	if err != nil {
		fmt.Println(err)
	}
}
