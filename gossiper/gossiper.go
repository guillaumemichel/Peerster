package gossiper

import (
	"log"
	"net"
)

// Gossiper : a gossiper
type Gossiper struct {
	Name    string
	address *net.UDPAddr
	conn    *net.UDPConn
}

// NewGossiper : creates a new gossiper with the given parameters
func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Panic(err)
	}
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Panic(err)
	}
	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name,
	}
}

// Start : starts a given gossiper
func (g *Gossiper) Start() {

}
