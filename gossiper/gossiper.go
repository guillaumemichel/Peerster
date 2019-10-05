package gossiper

import (
	"fmt"
	"log"
	"net"

	protobuf "github.com/DeDiS/protobuf"
	p "github.com/guillaumemichel/Peerster/peers"
	t "github.com/guillaumemichel/Peerster/types"
)

// Gossiper : a gossiper
type Gossiper struct {
	Name    string
	Address *net.UDPAddr
	Conn    *net.UDPConn
	Peers   *p.PeerList
}

// ErrorCheck : check for non critical error, and logs the result
func ErrorCheck(err error) {
	if err != nil {
		log.Println(err)
	}
}

// NewGossiper : creates a new gossiper with the given parameters
func NewGossiper(address, name, UIPort *string, peerList *p.PeerList) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", *address)
	if err != nil {
		log.Panic(err)
	} /// TODOOOOOOOOOOOOO MANAGE UIPORT
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Panic(err)
	}
	return &Gossiper{
		Address: udpAddr,
		Conn:    udpConn,
		Name:    *name,
		Peers:   peerList,
	}
}

// PrintMessageClient : print messages from the client
func (g *Gossiper) PrintMessageClient(packet t.GossipPacket) {
	simpleM := packet.Simple

	fmt.Println("CLIENT MESSAGE ", simpleM.Contents)
}

// HandleMessage : handles a message on arrival
func (g *Gossiper) HandleMessage(rcvBytes []byte, m int, udpAddr *net.UDPAddr) {
	var rcvMsg t.GossipPacket

	err := protobuf.Decode(rcvBytes, rcvMsg)
	ErrorCheck(err)

	g.PrintMessageClient(rcvMsg)
}

// ListenClient : listen for new messages
func (g *Gossiper) ListenClient() {
	var rcvBytes []byte

	for {
		m, addr, err := g.Conn.ReadFromUDP(rcvBytes)
		ErrorCheck(err)
		// may be vulnerable to DOS from client, but fast otherwise
		go g.HandleMessage(rcvBytes, m, addr)
	}
}

// Run : runs a given gossiper
func (g *Gossiper) Run() {
	go g.ListenClient()
}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name, UIPort *string, peerList *p.PeerList) {
	NewGossiper(address, name, UIPort, peerList).Run()
}
