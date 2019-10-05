package gossiper

import (
	"fmt"
	"log"
	"net"
	"strconv"

	protobuf "github.com/DeDiS/protobuf"
	p "github.com/guillaumemichel/Peerster/peers"
	t "github.com/guillaumemichel/Peerster/types"
)

// Gossiper : a gossiper
type Gossiper struct {
	Name       string
	GossipAddr *net.UDPAddr
	ClientAddr *net.UDPAddr
	GossipConn *net.UDPConn
	ClientConn *net.UDPConn
	Peers      *p.PeerList
}

// ErrorCheck : check for non critical error, and logs the result
func ErrorCheck(err error) {
	if err != nil {
		log.Println(err)
	}
}

// PanicCheck : check for panic level errors, and logs the result
func PanicCheck(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// NewGossiper : creates a new gossiper with the given parameters
func NewGossiper(address, name, UIPort *string, peerList *p.PeerList) *Gossiper {

	// define gossip address and connection for the new gossiper
	gossAddr, err := net.ResolveUDPAddr("udp4", *address)
	PanicCheck(err)
	gossConn, err := net.ListenUDP("udp4", gossAddr)
	PanicCheck(err)

	// sanitize the uiport for new gossiper
	cliPort, err := strconv.Atoi(*UIPort)
	if err != nil || cliPort < 0 || cliPort > 65535 {
		log.Fatalf("Error: invalid port %s", *UIPort)
	}

	// define client address and connection for the new gossiper
	cliAddr := &net.UDPAddr{
		IP:   gossAddr.IP,
		Port: cliPort,
	}
	cliConn, err := net.ListenUDP("udp4", cliAddr)
	PanicCheck(err)

	return &Gossiper{
		Name:       *name,
		GossipAddr: gossAddr,
		ClientAddr: cliAddr,
		GossipConn: gossConn,
		ClientConn: cliConn,
		Peers:      peerList,
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
		m, addr, err := g.ClientConn.ReadFromUDP(rcvBytes)
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
