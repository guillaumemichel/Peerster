package gossiper

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

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
	MsgCount   int
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
		MsgCount:   0,
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
func (g *Gossiper) PrintMessageClient(text string) {
	fmt.Println("CLIENT MESSAGE", text)
	g.PrintPeers()
}

// PrintMessageGossip : print messages received from gossipers
func (g *Gossiper) PrintMessageGossip(pack *u.GossipPacket) {

	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		pack.Simple.OriginalName, pack.Simple.RelayPeerAddr,
		pack.Simple.Contents)
	g.PrintPeers()
}

// CreateSimpleMessage : create a gossip simple message from a string
func (g *Gossiper) CreateSimpleMessage(content string) *u.SimpleMessage {
	return &u.SimpleMessage{
		OriginalName:  g.Name,
		RelayPeerAddr: g.GossipAddr.String(),
		Contents:      content,
	}
}

/*
// ReplaceOriginalNameSimple : replaces the original name of a simple message
// with its own name
func (g *Gossiper) ReplaceOriginalNameSimple(msg *u.SimpleMessage) *u.SimpleMessage {
	msg.OriginalName = g.Name
	return msg
}*/

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
	if !strings.Contains(g.PeersToString(), addr.String()) {
		g.Peers = append(g.Peers, *addr)
	}
}

// Propagate : Sends a message to all known gossipers
func (g *Gossiper) Propagate(packet []byte, sender *net.UDPAddr) {
	for _, v := range g.Peers {
		if !u.EqualAddr(&v, sender) {
			g.GossipConn.WriteToUDP(packet, &v)
		}
	}
}

// HandleMessage1 : handles a message on arrival
func (g *Gossiper) HandleMessage1(rcvBytes []byte, udpAddr *net.UDPAddr,
	mode string) {
	rcvMsg, ok := u.UnprotobufGossip(rcvBytes)

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
			packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
			g.Propagate(packet, udpAddr)
		case "client":
			// print the message in the console
			g.PrintMessageClient(rcvMsg.Simple.Contents)
			// update the SimpleMessage
			//sm = g.ReplaceOriginalNameSimple(sm)
			sm = g.ReplaceRelayPeerSimple(sm)
			// serialize the message
			packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
			// broadcast the message to all hosts
			g.Propagate(packet, nil)
		default:
			log.Fatal("Invalid message")
		}
	}

}

// ReceiveOK : return false if message too long to be handled, true otherwise
func (g *Gossiper) ReceiveOK(ok bool, rcvBytes []byte) bool {
	if !ok && len(rcvBytes) >= g.BufSize {
		log.Printf(`Warning: incoming message possibly larger than %d bytes 
			couldn't be read!`, g.BufSize)
		return false
	}
	return true
}

// HandleMessage : handles a message on arrival
func (g *Gossiper) HandleMessage(rcvBytes []byte, udpAddr *net.UDPAddr,
	gossip bool) {

	if gossip {
		rcvMsg, ok := u.UnprotobufGossip(rcvBytes)
		if g.ReceiveOK(ok, rcvBytes) {
			// TODO: handle different types of messages
			if g.MsgCount < 100 {

				sm := rcvMsg.Simple
				if udpAddr.String() != sm.RelayPeerAddr {
					println("Warning: relay peer address and sender address do not match")
				}
				// add the sender to known peers
				g.AddPeer(udpAddr)
				// prints message to console
				g.PrintMessageGossip(rcvMsg)
				// replace relay peer address by its own
				sm = g.ReplaceRelayPeerSimple(sm)
				// protobuf the new message
				packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
				// broadcast it, except to sender
				g.Propagate(packet, udpAddr)
				g.MsgCount++
			}
		}

	} else {
		rcvMsg, ok := u.UnprotobufMessage(rcvBytes)
		if g.ReceiveOK(ok, rcvBytes) {
			m := rcvMsg.Text
			// prints message to console
			g.PrintMessageClient(m)
			// creates a SimpleMessage to be broadcasted
			sm := g.CreateSimpleMessage(m)
			// protobuf the message
			packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
			// broadcast the message to all peers
			g.Propagate(packet, nil)
		}
	}
}

// Listen : listen for new messages from clients
func (g *Gossiper) Listen(udpConn *net.UDPConn) {
	buf := make([]byte, g.BufSize)
	// determine if we listen for Gossips or Client messages
	gossip := udpConn == g.GossipConn

	for {
		// read new message
		m, addr, err := udpConn.ReadFromUDP(buf)
		ErrorCheck(err)
		// whenever a new message arrives, start a new go routine to handle it
		go g.HandleMessage(buf[:m], addr, gossip)
	}
}

// Run : runs a given gossiper
func (g *Gossiper) Run() {

	// starts a listener on ui and gossip ports
	go g.Listen(g.ClientConn)
	go g.Listen(g.GossipConn)

	// keep the program active
	for {
	}
}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name, UIPort, peerList *string) {
	NewGossiper(address, name, UIPort, peerList).Run()
}
