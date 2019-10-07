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
	Mode       string
	RumorCount uint32
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
func NewGossiper(address, name, UIPort, peerList *string,
	simple bool) *Gossiper {

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
	var mode string
	if simple {
		mode = u.SimpleModeStr
	} else {
		mode = u.RumorModeStr
	}

	return &Gossiper{
		Name:       *name,
		GossipAddr: gossAddr,
		ClientAddr: cliAddr,
		GossipConn: gossConn,
		ClientConn: cliConn,
		Peers:      *peers,
		BufSize:    bufferSize,
		Mode:       mode,
		RumorCount: 0,
	}
}

// PrintPeers : print the known peers from the gossiper
func (g *Gossiper) PrintPeers() {
	fmt.Println("PEERS", g.PeersToString())
}

// PeersToString : return a string containing the list of known peers
func (g *Gossiper) PeersToString() *string {
	str := ""
	// if not peers return empty string
	if len(g.Peers) == 0 {
		return &str
	}
	for _, v := range g.Peers {
		str += v.String() + ","
	}
	// don't return the last ","
	str = str[:len(str)-1]
	return &str
}

// PrintMessageClient : print messages from the client
func (g *Gossiper) PrintMessageClient(text *string) {
	fmt.Println("CLIENT MESSAGE", *text)
	g.PrintPeers()
}

// PrintSimpleMessage : print simple messages received from gossipers
func (g *Gossiper) PrintSimpleMessage(msg *u.SimpleMessage, from *string) {

	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		msg.OriginalName, *from, msg.Contents)
	g.PrintPeers()
}

// PrintRumorMessage : print rumor messages received from gossipers
func (g *Gossiper) PrintRumorMessage(msg *u.RumorMessage, from *string) {
	fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n",
		msg.Origin, *from, msg.ID, msg.Text)
	g.PrintPeers()
}

// PrintStatusMessage : print status messages received from gossipers
func (g *Gossiper) PrintStatusMessage(msg *u.StatusPacket, from *string) {
	fmt.Printf("STATUS from %s", *from)
	for _, v := range msg.Want {
		fmt.Printf(" peer %s nextID %d", v.Identifier, v.NextID)
	}
	fmt.Println()
	g.PrintPeers()
}

// CreateSimpleMessage : creates a gossip simple message from a string
func (g *Gossiper) CreateSimpleMessage(name, content *string) *u.SimpleMessage {
	return &u.SimpleMessage{
		OriginalName:  *name,
		RelayPeerAddr: g.GossipAddr.String(),
		Contents:      *content,
	}
}

// CreateRumorMessage :creates a gossip rumor message from a string, and
// increases the rumorCount from the Gossiper
func (g *Gossiper) CreateRumorMessage(name, content *string) *u.RumorMessage {
	g.RumorCount++
	return &u.RumorMessage{
		Origin: g.Name,
		ID:     g.RumorCount,
		Text:   *content,
	}
}

// ReplaceRelayPeerSimple : replaces the relay peer of a simple message with its
// own address
func (g *Gossiper) ReplaceRelayPeerSimple(
	msg *u.SimpleMessage) *u.SimpleMessage {

	msg.RelayPeerAddr = g.GossipAddr.String()
	if msg.RelayPeerAddr == "<nil>" {
		log.Fatal("cannot replace relay peer address")
	}
	return msg
}

// AddPeer : adds the given peer to peers list if not already in it
func (g *Gossiper) AddPeer(addr *net.UDPAddr) {
	if !strings.Contains(*g.PeersToString(), addr.String()) {
		g.Peers = append(g.Peers, *addr)
	}
}

// Broadcast : Sends a message to all known gossipers
func (g *Gossiper) Broadcast(packet []byte, sender *net.UDPAddr) {
	for _, v := range g.Peers {
		if !u.EqualAddr(&v, sender) {
			g.GossipConn.WriteToUDP(packet, &v)
		}
	}
}

// SendToRandomPeer : sends the given packet to a random peer
func (g *Gossiper) SendToRandomPeer(packet []byte, peers *[]net.UDPAddr) {
	// get a random host to send the message
	r := u.GetRealRand(len(*peers)) // GetRand(n) also possible
	g.GossipConn.WriteToUDP(packet, &(*peers)[r])
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

	if gossip { // gossip message
		rcvMsg, ok := u.UnprotobufGossip(rcvBytes)
		if g.ReceiveOK(ok, rcvBytes) && u.TestMessageType(rcvMsg) {
			// add the sender to known peers
			g.AddPeer(udpAddr)
			addrStr := (*udpAddr).String()
			if g.Mode == u.SimpleModeStr { // SimpleMessage
				if rcvMsg.Simple != nil { // SimpleMessage received
					sm := rcvMsg.Simple
					// prints message to console
					g.PrintSimpleMessage(sm, &addrStr)
					// replace relay peer address by its own
					sm = g.ReplaceRelayPeerSimple(sm)
					// protobuf the new message
					packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
					// broadcast it, except to sender
					g.Broadcast(packet, udpAddr)
				} else if rcvMsg.Rumor != nil { // RumorMessage received
					// RumorMessage converted to SimpleMessage
					m := rcvMsg.Rumor
					// prints message to console
					g.PrintRumorMessage(m, &addrStr)
					fmt.Println("Warning: gossiper running in Simple mode",
						"and received a RumorMessage, discarding it")

					/*
						// replace the RumorMessage by a SimpleMessage
						sm := g.CreateSimpleMessage(&m.Origin, &m.Text)
						// protobuf the new message
						packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
						// broadcast it, except to sender
						g.Broadcast(packet, udpAddr)*/

				} else { // StatusMessage received
					m := rcvMsg.Status
					// prints message to console
					g.PrintStatusMessage(m, &addrStr)
					fmt.Println("Warning: gossiper running in Simple mode",
						"and received a StatusPacket, discarding it")
				}

			} else if g.Mode == u.RumorModeStr {
				if rcvMsg.Simple != nil { // SimpleMessage received
					sm := rcvMsg.Simple
					// prints message to console
					g.PrintSimpleMessage(sm, &addrStr)
					fmt.Println("Warning: gossiper running in Rumor mode",
						"and received a SimpleMessage, discarding it")
				} else if rcvMsg.Rumor != nil { // RumorMessage received
					// RumorMessage converted to SimpleMessage
					m := rcvMsg.Rumor
					// prints message to console
					g.PrintRumorMessage(m, &addrStr)
					// send it to a random peer, but not the sender
					possiblePeers := u.RemoveAddrFromPeers(&g.Peers, udpAddr)
					// as the message doesn't change, we send rcvBytes
					g.SendToRandomPeer(rcvBytes, possiblePeers)
				} else { // StatusMessage received
					m := rcvMsg.Status
					// prints message to console
					g.PrintStatusMessage(m, &addrStr)
					// TODO send request
				}
			}
		}

	} else { // message from the client
		rcvMsg, ok := u.UnprotobufMessage(rcvBytes)
		if g.ReceiveOK(ok, rcvBytes) {
			m := rcvMsg.Text
			// prints message to console
			g.PrintMessageClient(&m)

			if g.Mode == u.SimpleModeStr { // simple mode
				// creates a SimpleMessage in GossipPacket to be broadcasted
				gPacket := u.GossipPacket{Simple: g.CreateSimpleMessage(&m,
					&g.Name)}
				// protobuf the message
				packet := u.ProtobufGossip(&gPacket)
				// broadcast the message to all peers
				g.Broadcast(packet, nil)
			} else { // rumor mode
				// creates a RumorMessage in GossipPacket to be broadcasted
				gPacket := u.GossipPacket{Rumor: g.CreateRumorMessage(
					&g.Name, &m)}
				// protobuf the message
				packet := u.ProtobufGossip(&gPacket)
				// sends the packet to a random peer
				g.SendToRandomPeer(packet, &g.Peers)
			}
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
func StartNewGossiper(address, name, UIPort, peerList *string, simple bool) {
	NewGossiper(address, name, UIPort, peerList, simple).Run()
}
