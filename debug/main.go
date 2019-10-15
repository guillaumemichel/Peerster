package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	u "github.com/guillaumemichel/Peerster/utils"
)

func main() {
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000",
		"ip:port for the gossiper")
	msg := flag.String("msg", "", "message to be forged and send")
	id := flag.Int("id", 1, "id of the sent packet")
	origin := flag.String("origin", "Gossiper", "origin of the forged message")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	if *msg == "" {
		fmt.Println("Error: no message")
		os.Exit(1)
	}

	rumor := u.RumorMessage{
		Origin: *origin,
		ID:     uint32(*id),
		Text:   *msg,
	}

	gp := u.GossipPacket{Rumor: &rumor}
	packet := u.ProtobufGossip(&gp)

	address, err := net.ResolveUDPAddr("udp4", *gossipAddr)
	if err != nil {
		fmt.Println(err)
	}

	udpConn, err := net.DialUDP("udp4", nil, address)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	_, err = udpConn.Write(packet)
	if err != nil {
		fmt.Println(err)
	}

}
