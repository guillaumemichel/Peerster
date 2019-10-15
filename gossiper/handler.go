package gossiper

import (
	"fmt"
	"net"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// Broadcast : Sends a message to all known gossipers
func (g *Gossiper) Broadcast(packet []byte, sender *net.UDPAddr) {
	for _, v := range g.Peers {
		if !u.EqualAddr(&v, sender) {
			g.GossipConn.WriteToUDP(packet, &v)
		}
	}
}

// SendStatus : send status/ack to given peer
func (g *Gossiper) SendStatus(dst *net.UDPAddr) {

	sp := g.BuildStatusPacket()
	gossip := u.GossipPacket{Status: &sp}
	packet := u.ProtobufGossip(&gossip)

	g.GossipConn.WriteToUDP(packet, dst)
}

// SendRumor : send rumor to the given peer, deals with timeouts and all
func (g *Gossiper) SendRumor(packet []byte, rumor u.RumorMessage,
	addr net.UDPAddr, initial u.MessageReference) {

	if initial.Origin == "" || initial.ID < 1 {
		fmt.Println("Error: missing initial message in SendRumor")
		return
	}

	// protobuf the message
	if packet == nil || len(packet) == 0 {
		gPacket := u.GossipPacket{Rumor: &rumor}
		packet = u.ProtobufGossip(&gPacket)
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
	select {
	case <-timeout: // TIMEOUT
		g.PendingACKs.Delete(*pendingACKStr)
		// send the initial packet to a random peer
		packet := g.HistoryMessageToByte(initial)
		g.SendRumorToRandom(packet, rumor, initial)
		return
	case <-ackChan: // ACK
		g.PendingACKs.Delete(*pendingACKStr)
		return
	}
}

// SendRumorToRandom : sends a rumor with all specifications to a random peer
func (g *Gossiper) SendRumorToRandom(packet []byte,
	rumor u.RumorMessage, initial u.MessageReference) {

	// get a random host to send the message
	target := g.GetRandPeer()
	if target != nil {
		g.SendRumor(packet, rumor, *target, initial)
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
				g.PrintSimpleMessage(*sm, addrStr)
				// replace relay peer address by its own
				sm = g.ReplaceRelayPeerSimple(sm)
				// protobuf the new message
				packet := u.ProtobufGossip(&u.GossipPacket{Simple: sm})
				// broadcast it, except to sender
				g.Broadcast(packet, udpAddr)

			} else if rcvMsg.Rumor != nil { // RumorMessage received
				// prints message to console
				g.PrintRumorMessage(*(rcvMsg.Rumor), addrStr)
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a RumorMessage, discarding it")

			} else { // StatusMessage received
				m := rcvMsg.Status
				// prints message to console
				g.PrintStatusMessage(*m, addrStr)
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a StatusPacket, discarding it")
			}

		} else if g.Mode == u.RumorModeStr { // rumor mode
			if rcvMsg.Simple != nil { // SimpleMessage received
				// prints message to console
				g.PrintSimpleMessage(*(rcvMsg.Simple), addrStr)
				fmt.Println("Warning: gossiper running in Rumor mode",
					"and received a SimpleMessage, discarding it")

			} else if rcvMsg.Rumor != nil { // RumorMessage received
				rumor := rcvMsg.Rumor

				// write the message in history
				// and check if message already known
				// if message already known, discard it

				if g.WriteRumorToHistory(*rumor) {
					// prints message to console
					g.PrintRumorMessage(*rumor, addrStr)

					// ack the message
					g.SendStatus(udpAddr)

					// as the message doesn't change, we send rcvBytes
					ref := u.MessageReference{
						Origin: rumor.Origin,
						ID:     rumor.ID,
					}

					g.SendRumorToRandom(rcvBytes, *rumor, ref)
				}

			} else { // StatusMessage received
				m := rcvMsg.Status
				// prints message to console
				g.PrintStatusMessage(*m, addrStr)
				g.DealWithStatus(*m, addrStr, udpAddr)
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
		g.PrintMessageClient(m)

		if g.Mode == u.SimpleModeStr { // simple mode
			// creates a SimpleMessage in GossipPacket to be broadcasted
			sm := g.CreateSimpleMessage(g.Name, m)
			gPacket := u.GossipPacket{Simple: &sm}
			// protobuf the message
			packet := u.ProtobufGossip(&gPacket)
			// broadcast the message to all peers
			g.Broadcast(packet, nil)
		} else { // rumor mode

			// creates a RumorMessage in GossipPacket to be broadcasted
			rumor := g.CreateRumorMessage(m)
			gPacket := u.GossipPacket{Rumor: &rumor}

			//write message to history
			if g.WriteRumorToHistory(rumor) {
				// protobuf the message
				packet := u.ProtobufGossip(&gPacket)
				// sends the packet to a random peer
				ref := u.MessageReference{
					Origin: rumor.Origin,
					ID:     rumor.ID,
				}

				g.SendRumorToRandom(packet, rumor, ref)
			}
		}

	}
}

// DealWithStatus : deals with status messages
func (g *Gossiper) DealWithStatus(status u.StatusPacket, sender string,
	addr *net.UDPAddr) {

	var initialMessage u.MessageReference
	addrStr := (*addr).String()

	// is the status an acknowledgement packet ?
	ack := false

	// associate this ack with a pending one if any
	// needs to be done fast, so there will be a second similar loop with non-
	// critical operations
	for _, v := range status.Want {
		// if origin no in want list, add it
		if _, ok := g.WantList.Load(v.Identifier); !ok {
			g.WantList.Store(v.Identifier, uint32(1))
		}

		// acknowledge rumor with ID lower than the ack we just recieved
		for i := v.NextID; i > 0; i-- {
			// we look for an ID lower than v.NextID
			identifier := u.AckIdentifier{
				Peer:   sender,
				Origin: v.Identifier,
				ID:     i,
			}
			// if it is pending, we acknowledge it by writing to the channel
			if va, ok := g.PendingACKs.Load(identifier); ok {
				// write to the corresponding channel to stop timer in
				//rumor message
				va.(u.AckValues).Channel <- true
				// set the initial message to the first message acked
				initialMessage = va.(u.AckValues).InitialMessage
				ack = true
			}
		}
	}

	//check if peer wants packets that g have, and send them if any
	for _, v := range status.Want {
		// if v.NextID is lower than the message we want, then we have stored
		// the message that is wanted, so we send it to peer and return
		wantedID, _ := g.WantList.Load(v.Identifier)
		if v.NextID < wantedID.(uint32) {
			// reference of the message to recover from history
			ref := u.MessageReference{Origin: v.Identifier, ID: v.NextID}
			// rumor to send
			rumor := g.RecoverHistoryRumor(ref)
			if !ack {
				initialMessage = u.MessageReference{
					Origin: rumor.Origin,
					ID:     rumor.ID,
				}
			}
			// create the packet to send
			gp := u.GossipPacket{Rumor: &rumor}
			packet := u.ProtobufGossip(&gp)

			g.SendRumor(packet, rumor, *addr, initialMessage)
			return
		}
	}

	// iterate over g's wantlist, and look for files that peer doesn't have,
	// and send it
	f := func(k, v interface{}) bool {
		// g knows the name, but haven't received a message yet from the peer
		if v.(uint32) < 2 {
			return true
		}
		found := false
		identifier := k.(string)
		// check if we can find the identifier in the status packet
		for _, o := range status.Want {
			if o.Identifier == identifier {
				found = true
				break
			}
		}
		// if identifier not found in status, send 1st packet of that identifier
		if !found {
			array, ok := g.RumorHistory.Load(identifier)
			if ok && array.([]u.HistoryMessage)[0].ID == 1 {
				// create the rumor
				rumor := u.RumorMessage{
					Origin: k.(string),
					ID:     array.([]u.HistoryMessage)[0].ID,
					Text:   array.([]u.HistoryMessage)[0].Text,
				}
				// protobuf it
				gp := u.GossipPacket{Rumor: &rumor}
				packet := u.ProtobufGossip(&gp)

				// if status packet, define initial message
				if !ack {
					initialMessage = u.MessageReference{
						Origin: rumor.Origin,
						ID:     rumor.ID,
					}
				}
				// send it to the peer
				g.SendRumor(packet, rumor, *addr, initialMessage)
				return false
			}
		}
		return true
	}
	g.WantList.Range(f)

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
	g.PrintInSync(addrStr)

	// if ack message and 50% chance
	if ack && u.GetRealRand(2) == 0 {

		// select a random peer
		target := g.GetRandPeer()
		// print flipped coin message
		g.PrintFlippedCoin(target.String())

		// recover the initial message to send to a random peer
		rumor := g.RecoverHistoryRumor(initialMessage)

		// protobuf the message
		gPacket := u.GossipPacket{Rumor: &rumor}
		packet := u.ProtobufGossip(&gPacket)

		g.SendRumor(packet, rumor, *target, initialMessage)

	}
}
