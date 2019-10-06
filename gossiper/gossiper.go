package gossiper

import (
	"fmt"
	"log"
	"net"
	"strconv"

	u "github.com/guillaumemichel/Peerster/utils"
)

// TODO: client can edit bufferSize
var bufferSize int = 2048

// Gossiper : a gossiper
type Gossiper struct {
	Name       string
	GossipAddr *net.UDPAddr
	ClientAddr *net.UDPAddr
	GossipConn *net.UDPConn
	ClientConn *net.UDPConn
	Peers      []net.UDPAddr
	BufSize    int
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
func NewGossiper(address, name, UIPort, peerList *string) *Gossiper {

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

	peers := u.ParsePeers(peerList)

	return &Gossiper{
		Name:       *name,
		GossipAddr: gossAddr,
		ClientAddr: cliAddr,
		GossipConn: gossConn,
		ClientConn: cliConn,
		Peers:      *peers,
		BufSize:    bufferSize,
	}
}

// PrintPeers : print the known peers from the gossiper
func (g *Gossiper) PrintPeers() {
	fmt.Println("PEERS", g.PeersToString())
}

// PeersToString : return a string containing the list of known peers
func (g *Gossiper) PeersToString() string {
	str := ""
	// if not peers return empty string
	if len(g.Peers) == 0 {
		return str
	}
	for _, v := range g.Peers {
		str += v.String() + ","
	}
	// don't return the last ","
	return str[:len(str)-1]
}

// PrintMessageClient : print messages from the client
func (g *Gossiper) PrintMessageClient(packet *u.GossipPacket) {
	fmt.Println("CLIENT MESSAGE", packet.Simple.Contents)
	g.PrintPeers()
}

// PrintMessageGossip : print messages received from gossipers
func (g *Gossiper) PrintMessageGossip(pack *u.GossipPacket) {

	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		pack.Simple.OriginalName, pack.Simple.RelayPeerAddr,
		pack.Simple.Contents)
	g.PrintPeers()
}

// ReplaceOriginalNameSimple : replaces the original name of a simple message
// with its own name
func (g *Gossiper) ReplaceOriginalNameSimple(msg *u.SimpleMessage) *u.SimpleMessage {
	msg.OriginalName = g.Name
	return msg
}

// ReplaceRelayPeerSimple : replaces the relay peer of a simple message with its
// own address
func (g *Gossiper) ReplaceRelayPeerSimple(msg *u.SimpleMessage) *u.SimpleMessage {
	msg.RelayPeerAddr = g.GossipAddr.String()
	if msg.RelayPeerAddr == "<nil>" {
		log.Fatal("cannot replace relay peer address")
	}
	return msg
}

// AddPeer : adds the given peer to peers list if not already in it
func (g *Gossiper) AddPeer(addr *net.UDPAddr) {
	for _, v := range g.Peers {
		if u.EqualAddr(&v, addr) {
			return
		}
	}
	g.Peers = append(g.Peers, *addr)
}

// Propagate : Sends a message to all known gossipers
func (g *Gossiper) Propagate(packet []byte, sender *net.UDPAddr) {
	for _, v := range g.Peers {
		if !u.EqualAddr(&v, sender) {
			g.GossipConn.WriteToUDP(packet, &v)
		}
	}
}

// HandleMessage : handles a message on arrival
func (g *Gossiper) HandleMessage(rcvBytes []byte, udpAddr *net.UDPAddr,
	mode string) {
	rcvMsg, ok := u.UnprotobufMessage(rcvBytes)

	if !ok && len(rcvBytes) >= g.BufSize {
		log.Printf(`Warning: incoming message possibly larger than %d bytes 
			couldn't be read!`, g.BufSize)
	} else {
		sm := rcvMsg.Simple
		if udpAddr.String() != sm.RelayPeerAddr {
			println("Warning: relay peer address and sender address do not match")
		}

		switch mode {
		case "gossip":

			g.AddPeer(udpAddr)
			g.PrintMessageGossip(rcvMsg)
			sm = g.ReplaceRelayPeerSimple(sm)
			packet := u.ProtobufMessage(&u.GossipPacket{Simple: sm})
			g.Propagate(packet, udpAddr)
		case "client":
			// print the message in the console
			g.PrintMessageClient(rcvMsg)
			// update the SimpleMessage
			sm = g.ReplaceOriginalNameSimple(sm)
			sm = g.ReplaceRelayPeerSimple(sm)
			// serialize the message
			packet := u.ProtobufMessage(&u.GossipPacket{Simple: sm})
			// broadcast the message to all hosts
			g.Propagate(packet, nil)
		default:
			log.Fatal("Invalid message")
		}
	}

}

// Listen : listen for new messages from clients
func (g *Gossiper) Listen(udpConn *net.UDPConn) {
	buf := make([]byte, g.BufSize)
	var mode string

	for {
		m, addr, err := udpConn.ReadFromUDP(buf)
		ErrorCheck(err)

		if udpConn == g.GossipConn {
			mode = "gossip"
		} else if udpConn == g.ClientConn {
			mode = "client"
		} else {
			mode = "unknown"
		}

		go g.HandleMessage(buf[:m], addr, mode)
	}
}

// Run : runs a given gossiper
func (g *Gossiper) Run() {
	go g.Listen(g.ClientConn)
	go g.Listen(g.GossipConn)

	for {
	}
}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name, UIPort, peerList *string) {
	NewGossiper(address, name, UIPort, peerList).Run()
}
