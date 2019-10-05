package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	protobuf "github.com/DeDiS/protobuf"
	t "github.com/guillaumemichel/Peerster/types"
)

func main() {
	// treating the flags
	UIPort := flag.String("UIPort", "8080", "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	name := "Client"
	destination := "127.0.0.1"

	// generating simple message in gossip packet
	simpleM := t.SimpleMessage{
		OriginalName:  name,
		RelayPeerAddr: name,
		Contents:      *msg,
	}

	packetToSend := t.GossipPacket{Simple: &simpleM}

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
	udpConn, err := net.ListenUDP("udp4", &address)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// serializing the packet to send
	bytesToSend, err := protobuf.Encode(&packetToSend)

	// sending the packet over udp
	udpConn.WriteToUDP(bytesToSend, &address)
}
