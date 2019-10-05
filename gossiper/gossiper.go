package gossiper

import (
	"log"
	"net"

	p "github.com/guillaumemichel/Peerster/peers"
)

// Gossiper : a gossiper
type Gossiper struct {
	Name    string
	Address *net.UDPAddr
	Conn    *net.UDPConn
	Peers   *p.PeerList
}

// NewGossiper : creates a new gossiper with the given parameters
func NewGossiper(address, name *string, peerList *p.PeerList) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", *address)
	if err != nil {
		log.Panic(err)
	}
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

// Run : runs a given gossiper
func (g *Gossiper) Run() {

}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name *string, peerList *p.PeerList) {
	NewGossiper(address, name, peerList).Run()
}
