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
	UIPort := flag.String("UIPort", u.DefaultUIPort, "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; "+
		"can be omitted")
	file := flag.String("file", "", "file to be indexed by the gossiper")

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

	// create the packet to send
	packetToSend := u.Message{
		Text:        *msg,
		Destination: dest,
		File:        file,
	}

	// serializing the packet to send
	bytesToSend := u.ProtobufMessage(&packetToSend)

	// sending the packet over udp
	_, err = udpConn.Write(bytesToSend)
	if err != nil {
		fmt.Println(err)
	}
}
