package gossiper

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// TODO: atomicity

// TODO: client can edit bufferSize
var bufferSize int = 2048

// Gossiper : a gossiper
type Gossiper struct {
	Name         string
	GossipAddr   *net.UDPAddr
	ClientAddr   *net.UDPAddr
	GossipConn   *net.UDPConn
	ClientConn   *net.UDPConn
	Peers        []net.UDPAddr
	BufSize      int
	Mode         string
	RumorCount   uint32
	PendingACKs  map[u.AckIdentifier]u.AckValues
	WantList     map[string]uint32
	RumorHistory map[string][]u.HistoryMessage
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
	acks := make(map[u.AckIdentifier]u.AckValues)

	status := make(map[string]uint32)
	status[*name] = 1
	history := make(map[string][]u.HistoryMessage)

	return &Gossiper{
		Name:         *name,
		GossipAddr:   gossAddr,
		ClientAddr:   cliAddr,
		GossipConn:   gossConn,
		ClientConn:   cliConn,
		Peers:        *peers,
		BufSize:      bufferSize,
		Mode:         mode,
		RumorCount:   0,
		PendingACKs:  acks,
		WantList:     status,
		RumorHistory: history,
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

// PrintFlippedCoin : prints flipped coin message
func (g *Gossiper) PrintFlippedCoin(addr *string) {
	fmt.Printf("FLIPPED COIN sending rumor to %s\n", *addr)
}

// PrintInSync : prints in sync message
func (g *Gossiper) PrintInSync(addr *string) {
	fmt.Printf("IN SYNC WITH %s\n", *addr)
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

// WriteRumorToHistory : write a given message to Gossiper message history
func (g *Gossiper) WriteRumorToHistory(rumor *u.RumorMessage) {
	origin := rumor.Origin
	ID := rumor.ID
	text := rumor.Text
	for _, v := range g.RumorHistory[origin] {
		if v.ID == ID {
			// message already written to history
			return
		}
	}
	// write message to history
	if int(ID) < len(g.RumorHistory[origin]) { // packet received not in order
		// more recent messages have been received, but message missing

		// fill a missing slot
		g.RumorHistory[origin][ID-1] = u.HistoryMessage{ID: ID, Text: text}

		found := false
		for i, v := range g.RumorHistory[origin] {
			// check for empty slots in message history to define the wantlist
			if v.ID == 0 {
				g.WantList[origin] = uint32(i + 1)
				found = true
				break
			}
		}
		// if not, the last slot has been filled, set wantlist value to the end
		// of the history (next value)
		if !found {
			g.WantList[origin] = uint32(len(g.RumorHistory[origin]) + 1)
		}
	} else if ID == g.WantList[origin] { // the next packet that g doesn't have
		g.RumorHistory[origin] = append(g.RumorHistory[origin],
			u.HistoryMessage{ID: ID, Text: text})
		// update the wantlist to the next message
		g.WantList[origin]++
	} else { // not the wanted message, but a newer one
		// e.g waiting for message 2, and get message 4
		for i := g.WantList[origin]; i < ID; i++ {
			// create empty slots for missing message
			g.RumorHistory[origin] = append(g.RumorHistory[origin],
				u.HistoryMessage{ID: 0, Text: ""})
		}
		g.RumorHistory[origin] = append(g.RumorHistory[origin],
			u.HistoryMessage{ID: ID, Text: text})
		// don't update the wantlist, as wanted message still missing
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

// ReceiveOK : return false if message too long to be handled, true otherwise
func (g *Gossiper) ReceiveOK(ok bool, rcvBytes []byte) bool {
	if !ok && len(rcvBytes) >= g.BufSize {
		log.Printf(`Warning: incoming message possibly larger than %d bytes 
			couldn't be read!\n`, g.BufSize)
		return false
	}
	return true
}

// HistoryMessageToByte : recover a rumor message from history, protobuf it
// and sends back an array of bytes
func (g *Gossiper) HistoryMessageToByte(ref *u.MessageReference) *[]byte {
	// if ID ref is larger than array size or message not in array
	if len(g.RumorHistory[ref.Origin]) < int(ref.ID) ||
		g.RumorHistory[ref.Origin][ref.ID].ID != ref.ID {

		log.Println("Error: message queried and not found in history!")
		return nil
	}
	// recover the text in history to create the rumor
	rumor := u.RumorMessage{
		Origin: ref.Origin,
		ID:     ref.ID,
		Text:   g.RumorHistory[ref.Origin][ref.ID].Text,
	}
	// protobuf the rumor to get a byte array
	packet := u.ProtobufGossip(&u.GossipPacket{Rumor: &rumor})
	return &packet
}

// SendRumor : send rumor to the given peer, deals with timeouts and all
func (g *Gossiper) SendRumor(packet *[]byte, rumor *u.RumorMessage,
	addr *net.UDPAddr, initial *u.MessageReference) {

	targetStr := (*addr).String()
	// initialize the initial message if it is nil
	if initial == nil {
		ref := u.MessageReference{
			Origin: rumor.Origin,
			ID:     rumor.ID,
		}
		initial = &ref
	}

	// create a unique identifier for the message
	pendingACKStr := u.GetACKIdentifierSend(rumor, &targetStr)

	// associate a channel and initial message with unique message identifier
	// in Gossiper
	g.PendingACKs[*pendingACKStr] = u.AckValues{
		Channel:        make(chan bool),
		InitialMessage: initial,
	}

	fmt.Printf("MONGERING with %s\n", targetStr)
	// send packet
	g.GossipConn.WriteToUDP(*packet, addr)

	// creates the timeout
	timeout := make(chan bool)
	go func() {
		// timeout value defined in utils/constants.go
		time.Sleep(time.Duration(u.TimeoutValue) * time.Second)
		timeout <- true
	}()

	// TODO check this
	select {
	case <-timeout: // TIMEOUT
		delete(g.PendingACKs, *pendingACKStr)
		// send the initial packet to a random peer
		packet := g.HistoryMessageToByte(initial)
		if packet != nil {
			g.SendRumorToRandom(packet, rumor, initial)
		}
		return
	case <-g.PendingACKs[*pendingACKStr].Channel: // ACK
		delete(g.PendingACKs, *pendingACKStr)
		return
	}
}

// SendRumorToRandom : sends a rumor with all specifications to a random peer
func (g *Gossiper) SendRumorToRandom(packet *[]byte,
	rumor *u.RumorMessage, initial *u.MessageReference) {

	// get a random host to send the message
	r := u.GetRealRand(len(g.Peers)) // GetRand(n) also possible
	target := (g.Peers)[r]

	g.SendRumor(packet, rumor, &target, initial)
}

// SendStoredMessage : send a stored message to the given peer
func (g *Gossiper) SendStoredMessage(origin *string, id uint32,
	udpAddr net.UDPAddr) {

	gossip := u.GossipPacket{Rumor: &u.RumorMessage{
		Origin: *origin,
		ID:     id,
		Text:   g.RumorHistory[*origin][id-1].Text, // -1 ids start at 1
	}}
	fmt.Println(gossip)
	//packet := u.ProtobufGossip(&gossip)
	// TODO send packet
}

// DealWithStatus : deals with status messages
func (g *Gossiper) DealWithStatus(status *u.StatusPacket, sender *string) {

	var initialMessage u.MessageReference

	// associate this ack with a pending one if any
	// needs to be done fast, so there will be a second similar loop with non-
	// critical operations
	for _, v := range status.Want {
		// get the string which identifies the pending ACK
		pendingACKStr := u.GetACKIdentifierReceive(v.NextID,
			&v.Identifier, sender)
		// check if the ACK is pending
		if _, ok := g.PendingACKs[*pendingACKStr]; ok { // pending ACK found
			// write to the corresponding channel to stop timer in rumor message

			// TODO : check ack newer than sent packet
			g.PendingACKs[*pendingACKStr].Channel <- true
			initialMessage = *g.PendingACKs[*pendingACKStr].InitialMessage
			break
		}
	}

	// g and peer synchronized ?
	sync := true

	//check if peer wants packets that g have, and send them if any
	for _, v := range status.Want {
		if v.NextID < g.WantList[v.Identifier] {

		}
	}

	// check if g is late on peer, and request messages if true
	for _, v := range status.Want {
		// if origin no in want list, add it
		if _, ok := g.WantList[v.Identifier]; !ok {
			g.WantList[v.Identifier] = 1
		}
		if v.NextID > g.WantList[v.Identifier] {
			// request message
		}
	}

	if sync {
		fmt.Println(initialMessage)
	}

	// add new Origins (if any) to status
	// compare own status with received one (together with above)

	// if peer need packets -> send one to it

	// else if g is late, sends status for update

	// print sync message, 1/2 chance to send first packet to new peer

	// TODO this
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
					// prints message to console
					g.PrintRumorMessage(rcvMsg.Rumor, &addrStr)
					fmt.Println("Warning: gossiper running in Simple mode",
						"and received a RumorMessage, discarding it")

				} else { // StatusMessage received
					m := rcvMsg.Status
					// prints message to console
					g.PrintStatusMessage(m, &addrStr)
					fmt.Println("Warning: gossiper running in Simple mode",
						"and received a StatusPacket, discarding it")
				}

			} else if g.Mode == u.RumorModeStr {
				if rcvMsg.Simple != nil { // SimpleMessage received
					// prints message to console
					g.PrintSimpleMessage(rcvMsg.Simple, &addrStr)
					fmt.Println("Warning: gossiper running in Rumor mode",
						"and received a SimpleMessage, discarding it")

				} else if rcvMsg.Rumor != nil { // RumorMessage received
					rumor := rcvMsg.Rumor
					// prints message to console
					g.PrintRumorMessage(rumor, &addrStr)
					// as the message doesn't change, we send rcvBytes
					g.SendRumorToRandom(&rcvBytes, rumor, nil)
					g.WriteRumorToHistory(rumor)

				} else { // StatusMessage received
					m := rcvMsg.Status
					// prints message to console
					g.PrintStatusMessage(m, &addrStr)
					g.DealWithStatus(m, &addrStr)
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
				rumor := g.CreateRumorMessage(&g.Name, &m)
				gPacket := u.GossipPacket{Rumor: rumor}
				// protobuf the message
				packet := u.ProtobufGossip(&gPacket)
				// sends the packet to a random peer
				g.SendRumorToRandom(&packet, rumor, nil)
				g.WriteRumorToHistory(rumor)
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
