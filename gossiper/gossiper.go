package gossiper

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
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
	PendingACKs  sync.Map // map[u.AckIdentifier]u.AckValues
	WantList     sync.Map // map[string]uint32
	RumorHistory sync.Map // map[string][]u.HistoryMessage
	AntiEntropy  int
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
	simple bool, antiE int) *Gossiper {

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
	var acks sync.Map
	var history sync.Map
	var status sync.Map

	status.Store(*name, uint32(1))

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
		AntiEntropy:  antiE,
	}
}

// PrintPeers : print the known peers from the gossiper
func (g *Gossiper) PrintPeers() {
	fmt.Println("PEERS", *(g.PeersToString()))
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
// return false if message already known, true otherwise
func (g *Gossiper) WriteRumorToHistory(rumor *u.RumorMessage) bool {
	origin := rumor.Origin
	ID := rumor.ID
	text := rumor.Text

	if originHistory, ok := g.RumorHistory.Load(origin); !ok {
		var empty []u.HistoryMessage
		g.RumorHistory.Store(origin, empty)
		g.WantList.Store(origin, uint32(1))
	} else {
		for _, v := range originHistory.([]u.HistoryMessage) {
			if v.ID == ID {
				// message already written to history
				// DEBUG
				//fmt.Println("MESSAGE ALREADY WRITTEN, DISCARD")
				return false
			}
		}
	}

	// write message to history
	oH, ok := g.RumorHistory.Load(origin)
	originHistory := oH.([]u.HistoryMessage)
	if !ok {
		fmt.Println("DEBUG: not ok, write rumor to history")
	}
	if int(ID) < len(originHistory) { // packet received not in order
		// more recent messages have been received, but message missing

		// fill a missing slot
		originHistory[ID-1] =
			u.HistoryMessage{ID: ID, Text: text}

		found := false
		for i, v := range originHistory {
			// check for empty slots in message history to define the wantlist
			if v.ID == 0 {
				g.WantList.Store(origin, uint32(i+1))
				found = true
				break
			}
		}
		// if not, the last slot has been filled, set wantlist value to the end
		// of the history (next value)
		if !found {
			g.WantList.Store(origin, uint32(len(originHistory)+1))
		}
	} else if newID, _ := g.WantList.Load(origin); ID == newID {
		// the next packet that g doesn't have
		originHistory = append(originHistory,
			u.HistoryMessage{ID: ID, Text: text})
		g.RumorHistory.Store(origin, originHistory)

		// NOOOOOOOO MUTEX CAN BE SHITTY AROUND HERE

		// update the wantlist to the next message
		curr, _ := g.WantList.Load(origin)
		g.WantList.Store(origin, uint32(curr.(uint32)+1))
	} else { // not the wanted message, but a newer one
		// e.g waiting for message 2, and get message 4
		for i, _ := g.WantList.Load(origin); i.(uint32) < ID; i = i.(uint32) + 1 {
			// create empty slots for missing message
			originHistory = append(originHistory,
				u.HistoryMessage{ID: 0, Text: ""})
		}
		originHistory = append(originHistory,
			u.HistoryMessage{ID: ID, Text: text})
		g.RumorHistory.Store(origin, originHistory)
		// don't update the wantlist, as wanted message still missing
	}
	return true
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

// RecoverHistoryRumor : recover a message from rumor history
func (g *Gossiper) RecoverHistoryRumor(ref u.MessageReference) u.RumorMessage {

	oH, ok := g.RumorHistory.Load(ref.Origin)
	//fmt.Printf("DEBUG: history size: %d, query: %d\n",len(originHistory), ref.ID)
	// if ID ref is larger than array size or message not in array
	if ok || len(oH.([]u.HistoryMessage)) < int(ref.ID) ||
		oH.([]u.HistoryMessage)[ref.ID-1].ID != ref.ID {

		log.Println("Error: message queried and not found in history!")
		return u.RumorMessage{}
	}

	// recover the text in history to create the rumor
	rumor := u.RumorMessage{
		Origin: ref.Origin,
		ID:     ref.ID,
		Text:   oH.([]u.HistoryMessage)[ref.ID-1].Text,
	}
	return rumor
}

// HistoryMessageToByte : recover a rumor message from history, protobuf it
// and sends back an array of bytes
func (g *Gossiper) HistoryMessageToByte(ref u.MessageReference) []byte {
	rumor := g.RecoverHistoryRumor(ref)
	// protobuf the rumor to get a byte array
	packet := u.ProtobufGossip(&u.GossipPacket{Rumor: &rumor})
	return packet
}

// SendRumor : send rumor to the given peer, deals with timeouts and all
func (g *Gossiper) SendRumor(packet []byte, rumor u.RumorMessage,
	addr net.UDPAddr, initial u.MessageReference) {

	// protobuf the message
	if len(packet) == 0 {
		gPacket := u.GossipPacket{Rumor: &rumor}
		p := u.ProtobufGossip(&gPacket)
		packet = p
	}

	targetStr := addr.String()

	// create a unique identifier for the message
	pendingACKStr := u.GetACKIdentifierSend(&rumor, &targetStr)

	// associate a channel and initial message with unique message identifier
	// in Gossiper
	values := u.AckValues{
		Channel:        make(chan bool),
		InitialMessage: initial,
	}
	g.PendingACKs.Store(*pendingACKStr, values)

	fmt.Printf("MONGERING with %s\n", targetStr)
	// send packet
	g.GossipConn.WriteToUDP(packet, &addr)

	// creates the timeout
	timeout := make(chan bool)
	go func() {
		// timeout value defined in utils/constants.go
		time.Sleep(time.Duration(u.TimeoutValue) * time.Second)
		timeout <- true
	}()

	ackChan := values.Channel
	// TODO check this
	select {
	case <-timeout: // TIMEOUT
		g.PendingACKs.Delete(*pendingACKStr)
		// send the initial packet to a random peer
		packet := g.HistoryMessageToByte(initial)
		if packet != nil {
			g.SendRumorToRandom(packet, rumor, initial, false)
		}
		return
	case <-ackChan: // ACK
		g.PendingACKs.Delete(*pendingACKStr)
		return
	}
}

// BuildStatusPacket : build a status packet for g
func (g *Gossiper) BuildStatusPacket() *u.StatusPacket {
	var want []u.PeerStatus

	f := func(k, v interface{}) bool {
		want = append(want, u.PeerStatus{Identifier: k.(string),
			NextID: v.(uint32)})
		return true
	}

	g.WantList.Range(f)
	sp := u.StatusPacket{Want: want}
	return &sp
}

// SendStatus : send status/ack to given peer
func (g *Gossiper) SendStatus(dst *net.UDPAddr) {
	// DEBUG
	// PrintWantlist(g)

	gossip := u.GossipPacket{Status: g.BuildStatusPacket()}
	packet := u.ProtobufGossip(&gossip)

	g.GossipConn.WriteToUDP(packet, dst)
}

// SendRumorToRandom : sends a rumor with all specifications to a random peer
func (g *Gossiper) SendRumorToRandom(packet []byte,
	rumor u.RumorMessage, initial u.MessageReference, coin bool) {

	// get a random host to send the message
	target := g.GetRandPeer()
	if target != nil {
		if coin {
			targetStr := (*target).String()
			g.PrintFlippedCoin(&targetStr)
		}

		g.SendRumor(packet, rumor, *target, initial)

	}
}

/*
// SendStoredMessage : send a stored message to the given peer
func (g *Gossiper) SendStoredMessage(origin *string, id uint32,
	udpAddr *net.UDPAddr) {

	gossip := u.GossipPacket{Rumor: &u.RumorMessage{
		Origin: *origin,
		ID:     id,
		Text:   g.RumorHistory[*origin][id-1].Text, // -1 ids start at 1
	}}

	packet := u.ProtobufGossip(&gossip)
	g.GossipConn.WriteToUDP(packet, udpAddr)
	//timout
}
*/

// DealWithStatus : deals with status messages
func (g *Gossiper) DealWithStatus(status *u.StatusPacket, sender *string,
	addr *net.UDPAddr) {

	var initialMessage u.MessageReference
	addrStr := (*addr).String()

	// is the status an acknowledgement packet ?
	ack := false

	// associate this ack with a pending one if any
	// needs to be done fast, so there will be a second similar loop with non-
	// critical operations
	for _, v := range status.Want {
		// get the string which identifies the pending ACK
		pendingACKId := u.GetACKIdentifierReceive(v.NextID,
			&v.Identifier, sender)

		// acknowledge rumor with ID lower than the ack we just recieved
		for i := v.NextID; i > 0; i-- {
			// we look for an ID lower than v.NextID
			identifier := u.AckIdentifier{
				Peer:   *sender,
				Origin: v.Identifier,
				ID:     i,
			}
			// if it is pending, we acknowledge it by writing to the channel
			if v, ok := g.PendingACKs.Load(identifier); ok {
				// write to the corresponding channel to stop timer in
				//rumor message
				v.(u.AckValues).Channel <- true
				// set the initial message to the first message acked
				fmt.Println("DEBUG: pending ack id", pendingACKId)
				initialMessage = v.(u.AckValues).InitialMessage
				ack = true
			}
		}
	}

	//check if peer wants packets that g have, and send them if any
	for _, v := range status.Want {
		// if origin no in want list, add it
		if _, ok := g.WantList.Load(v.Identifier); !ok {
			g.WantList.Store(v.Identifier, uint32(1))
		}

		// if v.NextID is lower than the message we want, then we have stored
		// the message that is wanted, so we send it to peer and return
		wantedID, _ := g.WantList.Load(v.Identifier)
		if v.NextID < wantedID.(uint32) {
			// reference of the message to recover from history
			ref := u.MessageReference{Origin: v.Identifier, ID: v.NextID}
			// rumor to send
			rumor := g.RecoverHistoryRumor(ref)
			g.SendRumor(nil, rumor, *addr, initialMessage)
			return
		}
	}

	// check if g is late on peer, and request messages if true
	for _, v := range status.Want {
		// if nextID > a message we want, request it
		wantedID, _ := g.WantList.Load(v.Identifier)
		if v.NextID > wantedID.(uint32) {
			// send status to request it
			g.SendStatus(addr)
			return
		}
	}

	// if we arrive here, then we are sync with peer

	// print in sync message
	g.PrintInSync(&addrStr)
	// if ack message and 50% chance
	if ack && u.GetRealRand(2) == 0 {
		// recover the initial message to send to a random peer
		rumor := g.RecoverHistoryRumor(initialMessage)

		g.SendRumorToRandom(nil, rumor, initialMessage, true)
	}
}

// HandleGossip : handle a gossip message
func (g *Gossiper) HandleGossip(rcvBytes []byte, udpAddr *net.UDPAddr) {
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

		} else if g.Mode == u.RumorModeStr { // rumor mode
			if rcvMsg.Simple != nil { // SimpleMessage received
				// prints message to console
				g.PrintSimpleMessage(rcvMsg.Simple, &addrStr)
				fmt.Println("Warning: gossiper running in Rumor mode",
					"and received a SimpleMessage, discarding it")

			} else if rcvMsg.Rumor != nil { // RumorMessage received
				rumor := rcvMsg.Rumor

				// write the message in history
				// and check if message already known
				// if message already known, discard it
				//DEBUG
				//fmt.Println("DEBUG: rcvd frm peer")

				if g.WriteRumorToHistory(rumor) {
					// prints message to console
					g.PrintRumorMessage(rumor, &addrStr)

					// ack the message
					g.SendStatus(udpAddr)

					// as the message doesn't change, we send rcvBytes
					ref := u.MessageReference{
						Origin: rumor.Origin,
						ID:     rumor.ID,
					}

					g.SendRumorToRandom(rcvBytes, *rumor, ref, false)
				}

			} else { // StatusMessage received
				m := rcvMsg.Status
				// prints message to console
				g.PrintStatusMessage(m, &addrStr)
				g.DealWithStatus(m, &addrStr, udpAddr)
			}
		}
	}
}

// HandleMessage : handles a message on arrival
func (g *Gossiper) HandleMessage(rcvBytes []byte, udpAddr *net.UDPAddr) {

	// message from the client
	rcvMsg, ok := u.UnprotobufMessage(rcvBytes)
	if g.ReceiveOK(ok, rcvBytes) {
		m := rcvMsg.Text
		// prints message to console
		g.PrintMessageClient(&m)

		if g.Mode == u.SimpleModeStr { // simple mode
			// creates a SimpleMessage in GossipPacket to be broadcasted
			gPacket := u.GossipPacket{Simple: g.CreateSimpleMessage(
				&g.Name, &m)}
			// protobuf the message
			packet := u.ProtobufGossip(&gPacket)
			// broadcast the message to all peers
			g.Broadcast(packet, nil)
		} else { // rumor mode

			// creates a RumorMessage in GossipPacket to be broadcasted
			rumor := g.CreateRumorMessage(&g.Name, &m)
			gPacket := u.GossipPacket{Rumor: rumor}

			//write message to history
			g.WriteRumorToHistory(rumor)

			// protobuf the message
			packet := u.ProtobufGossip(&gPacket)
			// sends the packet to a random peer
			ref := u.MessageReference{
				Origin: rumor.Origin,
				ID:     rumor.ID,
			}

			g.SendRumorToRandom(packet, *rumor, ref, false)
		}

	}
}

// Listen : listen for new messages from clients
func (g *Gossiper) Listen(udpConn *net.UDPConn) {
	buf := make([]byte, g.BufSize)

	for {
		// read new message
		m, addr, err := udpConn.ReadFromUDP(buf)
		ErrorCheck(err)

		tmp := make([]byte, m)
		copy(tmp, buf[:m])

		// whenever a new message arrives, start a new go routine to handle it
		if udpConn == g.GossipConn { // gossip
			go g.HandleGossip(tmp, addr)
		} else { // message from client
			go g.HandleMessage(tmp, addr)
		}
	}
}

// GetRandPeer : get a random peer known to g
func (g *Gossiper) GetRandPeer() *net.UDPAddr {
	if len(g.Peers) == 0 {
		fmt.Println("Error: cannot send message to random peer,",
			"no known peer :(")
		return nil
	}
	r := u.GetRealRand(len(g.Peers)) // GetRand(n) also possible
	target := (g.Peers)[r]
	return &target
}

// DoAntiEntropy : manage anti entropy
func (g *Gossiper) DoAntiEntropy() {
	// if anti entropy is 0 (or less) anti entropy disabled
	if g.AntiEntropy > 0 {
		for {
			// sleep for anti entropy value
			time.Sleep(time.Duration(g.AntiEntropy) * time.Second)

			if len(g.Peers) > 0 {
				// send status to random peer
				target := g.GetRandPeer()
				g.SendStatus(target)

			}
		}

	}
}

// Run : runs a given gossiper
func (g *Gossiper) Run() {

	// starts a listener on ui and gossip ports
	go g.Listen(g.ClientConn)
	go g.Listen(g.GossipConn)

	// keep the program active
	if g.Mode == u.RumorModeStr {
		g.DoAntiEntropy()
	}
	for {

	}
}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name, UIPort, peerList *string,
	simple bool, antiE int) {
	NewGossiper(address, name, UIPort, peerList, simple, antiE).Run()
}
