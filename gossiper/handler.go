package gossiper

import (
	"fmt"
	"net"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// Broadcast : Sends a message to all known gossipers
func (g *Gossiper) Broadcast(packet []byte, sender *net.UDPAddr) {
	//g.PeerMutex.Lock()
	for _, v := range g.Peers {
		if sender == nil || !u.EqualAddr(v, *sender) {
			g.GossipConn.WriteToUDP(packet, &v)
		}
	}
	//g.PeerMutex.Unlock()
}

// SendStatus : send status/ack to given peer
func (g *Gossiper) SendStatus(dst *net.UDPAddr) {

	sp := g.BuildStatusPacket()
	gossip := u.GossipPacket{Status: &sp}
	packet := u.ProtobufGossip(&gossip)

	g.GossipConn.WriteToUDP(packet, dst)
}

// DealWithPrivateMessage deals with private messages
func (g *Gossiper) DealWithPrivateMessage(pm u.PrivateMessage) {
	dst := pm.Destination
	// if the destination is g, print the private message to console
	if dst == g.Name {
		g.PrintPrivateMessage(pm)
		g.PrivateMsg = append(g.PrivateMsg, pm)
	} else {
		// decrease the hop limit
		pm.HopLimit--
		// if positive forward it, otherwise drop it
		if pm.HopLimit > 0 {
			gp := u.GossipPacket{Private: &pm}
			g.RoutePacket(dst, gp)
		}
	}
}

// RouteDataReq route data request to next hop
func (g *Gossiper) RouteDataReq(dreq u.DataRequest) {
	if g.ShouldPrint(logHW3, 3) {
		g.Printer.Println("I got a data request to route", dreq)
		g.Printer.Println("My name:", g.Name, "Dest:", dreq.Destination)
	}

	dst := dreq.Destination
	if dst == g.Name {
		g.HandleDataReq(dreq)
	} else {
		// decrease the hop limit
		dreq.HopLimit--
		// if positive forward it, otherwise drop it
		if dreq.HopLimit > 0 {
			gp := u.GossipPacket{DataRequest: &dreq}
			g.RoutePacket(dst, gp)
		}
	}
}

// RouteDataReply route data reply to next hop
func (g *Gossiper) RouteDataReply(drep u.DataReply) {
	dst := drep.Destination
	if dst == g.Name {
		g.HandleDataReply(drep)
	} else {
		// decrease the hop limit
		drep.HopLimit--
		// if positive forward it, otherwise drop it
		if drep.HopLimit > 0 {
			gp := u.GossipPacket{DataReply: &drep}
			g.RoutePacket(dst, gp)
		}
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
				rumor := *(rcvMsg.Rumor)
				if rumor.Text != "" {
					g.PrintRumorMessage(rumor, addrStr)
					fmt.Println("Warning: gossiper running in Simple mode",
						"and received a RumorMessage, discarding it")
				}

			} else if rcvMsg.Status != nil { // StatusMessage received
				m := rcvMsg.Status
				// prints message to console
				g.PrintStatusMessage(*m, addrStr)
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a StatusPacket, discarding it")

			} else if rcvMsg.Private != nil { // PrivateMessage received
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a Private message, discarding it")

			} else if rcvMsg.DataRequest != nil {
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a Data Request, discarding it")

			} else if rcvMsg.DataReply != nil {
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a Data Reply, discarding it")

			} else if rcvMsg.TLCMessage != nil {
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a TLC message, discarding it")

			} else if rcvMsg.Ack != nil {
				fmt.Println("Warning: gossiper running in Simple mode",
					"and received a TLC ack, discarding it")

			} else {
				fmt.Println("Error: unrecognized message")
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

				// update the route if message not already received
				g.UpdateRoute(*rumor, addrStr)

				// ack the message
				g.SendStatus(udpAddr)

				if g.WriteGossipToHistory(*rcvMsg) {
					// prints message to console
					g.PrintRumorMessage(*rumor, addrStr)

					// verify if we should send a TLC confirmation
					if g.Hw3ex3 {
						g.VerifyConfirmTLC()
					}

					// monger to random peer
					g.Monger(rcvMsg, rcvMsg, *g.GetRandPeer())
				}

			} else if rcvMsg.Status != nil {
				// StatusMessage received
				m := rcvMsg.Status
				// prints message to console
				g.PrintStatusMessage(*m, addrStr)
				g.HandleStatus(*m, *udpAddr)

			} else if rcvMsg.Private != nil {
				// PrivateMessage received
				pm := rcvMsg.Private
				g.DealWithPrivateMessage(*pm)

			} else if rcvMsg.DataRequest != nil {
				// deals with data request
				dr := rcvMsg.DataRequest
				g.RouteDataReq(*dr)

			} else if rcvMsg.DataReply != nil {
				// deals with data reply
				g.RouteDataReply(*rcvMsg.DataReply)

			} else if rcvMsg.SearchRequest != nil {
				// deals with search requests
				g.HandleSearchReq(*rcvMsg.SearchRequest, udpAddr)

			} else if rcvMsg.SearchReply != nil {
				// deals with search reply
				if g.ShouldPrint(logHW3, 3) {
					g.Printer.Println("Got search reply")
				}
				g.HandleSearchReply(*rcvMsg.SearchReply)

			} else if rcvMsg.TLCMessage != nil {
				// TLC message

				// send status to neighbor
				g.SendStatus(udpAddr)
				// verify if we should send tlc ack for previous tlc
				if g.Hw3ex3 {
					g.VerifyConfirmTLC()
				}
				// handle tlc
				g.HandleTLCMessage(*rcvMsg)

			} else if rcvMsg.Ack != nil {
				// TLC ack
				g.HandleTLCAcks(*rcvMsg.Ack)

			} else {
				g.Printer.Println(
					"Error: unrecognized message, forgot hw3 flag ?")
			}
		}
	} else {
		if g.ShouldPrint(logHW3, 2) {
			g.Printer.Println("Invalid message from :", *udpAddr)
		}
	}
}

// HandleMessage : handles a message on arrival
func (g *Gossiper) HandleMessage(rcvBytes []byte, udpAddr *net.UDPAddr) {
	// unprotobuf message from the client
	rcvMsg, ok := u.UnprotobufMessage(rcvBytes)
	if g.ReceiveOK(ok, rcvBytes) {
		if g.Mode != u.RumorModeStr && g.Mode != u.SimpleModeStr {
			g.PrintUnknownMode()
		}
		bText := rcvMsg.Text != ""
		bDest := rcvMsg.Destination != nil && *rcvMsg.Destination != ""
		bFile := rcvMsg.File != nil && *rcvMsg.File != ""
		bReq := rcvMsg.Request != nil && len(*rcvMsg.Request) > 0
		bKw := rcvMsg.Keywords != nil && len(*rcvMsg.Keywords) > 0

		if bText && !bDest && !bFile && !bReq && !bKw {
			// rumor message

			// print the client message status
			g.PrintMessageClient(rcvMsg.Text)
			if g.Mode == u.SimpleModeStr { // simple mode
				// creates a SimpleMessage in GossipPacket to be broadcasted
				sm := g.CreateSimpleMessage(g.Name, rcvMsg.Text)
				gPacket := u.GossipPacket{Simple: &sm}
				// protobuf the message
				packet := u.ProtobufGossip(&gPacket)
				// broadcast the message to all peers
				g.Broadcast(packet, nil)

			} else { // rumor mode
				// creates a RumorMessage in GossipPacket to be broadcasted
				rumor := g.CreateRumorMessage(rcvMsg.Text)
				gp := u.GossipPacket{Rumor: &rumor}

				//write message to history
				if g.WriteGossipToHistory(gp) {
					// monger message to random peer
					g.Monger(&gp, &gp, *g.GetRandPeer())
				}
			}

		} else if bText && bDest && !bFile && !bReq && !bKw {
			// private message

			if g.Mode == u.RumorModeStr {
				g.PrintSentPrivateMessage(*rcvMsg.Destination, rcvMsg.Text)
				// create private message
				pm := g.CreatePrivateMessage(rcvMsg.Text, *rcvMsg.Destination)
				// append the private message to the list of pms
				g.PrivateMsg = append(g.PrivateMsg, pm)
				// deals with it!
				g.DealWithPrivateMessage(pm)
			} else {
				g.PrintExpectedRumorMode("private message")
			}

		} else if !bText && !bDest && bFile && !bReq && !bKw {
			// index file

			// send metafile to host
			// g.SendFileTo(*rcvMsg.Destination, *rcvMsg.File)
			if g.Mode == u.RumorModeStr {
				g.IndexFile(*rcvMsg.File)
			} else {
				g.PrintExpectedRumorMode("file index")
			}

		} else if !bText && bDest && bFile && bReq && !bKw {
			// file request to host

			if g.Mode == u.RumorModeStr {
				// create a file request and sends it
				g.RequestFile(*rcvMsg.File,
					*rcvMsg.Destination, *rcvMsg.Request)
			} else {
				g.PrintExpectedRumorMode("file request")
			}

		} else if !bText && !bDest && bFile && bReq && !bKw {
			// file request after search
			g.HandleDownload(*rcvMsg.File, *rcvMsg.Request)

		} else if !bText && !bDest && !bFile && !bReq && bKw {
			// file search
			g.ManageSearch(rcvMsg.Budget, *rcvMsg.Keywords)

		} else {
			g.Printer.Println("Error: invalid message received")
		}
	}
}

// SendRouteRumor send a route rumor to a random peer
func (g *Gossiper) SendRouteRumor() {
	// create a route rumor message
	rumor := g.CreateRouteMessage()
	// write the route rumor in history

	gp := u.GossipPacket{Rumor: &rumor}
	if !g.WriteGossipToHistory(gp) {
		g.Printer.Println("Fatal: couldn't write own route rumor to history")
	}
	// sends it to a random peer
	g.Monger(&gp, &gp, *g.GetRandPeer())
}

// Monger TLC messages
func (g *Gossiper) Monger(gp, initial *u.GossipPacket, addr net.UDPAddr) {
	if gp == nil {
		return
	}

	addrStr := addr.String()

	var origin string
	var id uint32
	if gp.Rumor != nil {
		origin = gp.Rumor.Origin
		id = gp.Rumor.ID
		g.PrintMongering(addrStr)

	} else if gp.TLCMessage != nil {
		tlc := gp.TLCMessage
		//tx := tlc.TxBlock.Transaction

		origin = tlc.Origin
		id = tlc.ID
		/*
			if tlc.Confirmed >= 0 && tlc.Origin != g.Name {
				g.PrintConfirmedGossip(origin, tx.Name,
					hex.EncodeToString(tx.MetafileHash), int(id), int(tx.Size))
			} else if tlc.Confirmed < 0 {
				g.PrintUnconfirmedGossip(origin, tx.Name,
					hex.EncodeToString(tx.MetafileHash), int(id), int(tx.Size))
			}
		*/
	}

	pendingACKStr := u.GetAckIdentifier(origin, addrStr, id)

	c := make(chan bool)

	g.ACKMutex.Lock()
	// put the channel in BlockRumor
	g.PendingGossip[pendingACKStr] = u.AckValues{
		Channel:        &c,
		InitialMessage: *initial,
	}
	g.ACKMutex.Unlock()

	// protobuf and send gossip packet
	g.SendPacketToNeighbor(addr, *gp)

	// creates the timeout
	timeout := make(chan bool)
	go func() {
		// timeout value defined in utils/constants.go
		time.Sleep(time.Duration(u.TimeoutValue) * time.Second)
		timeout <- true
	}()

	select {
	case <-timeout: // TIMEOUT
		g.ACKMutex.Lock()
		delete(g.PendingGossip, pendingACKStr)
		g.ACKMutex.Unlock()
		// send the monger initial packet to a random peer
		g.Monger(initial, initial, *g.GetRandPeer())

	case <-c: // ACK
		g.ACKMutex.Lock()
		delete(g.PendingGossip, pendingACKStr)
		g.ACKMutex.Unlock()

	}
}

// HandleStatus handles a status received from a peer
func (g *Gossiper) HandleStatus(status u.StatusPacket, addr net.UDPAddr) {
	addrStr := addr.String()
	ack := false
	var initialPacket u.GossipPacket

	// associate this ack with a pending one if any
	// needs to be done fast, so there will be a second similar loop with non-
	// critical operations
	for _, v := range status.Want {
		// if origin no in want list, add it
		if _, ok := g.WantList[v.Identifier]; !ok {
			//g.WantListMutex.Lock()
			g.WantList[v.Identifier] = uint32(1)
			//g.WantListMutex.Unlock()
		}
		// acknowledge rumor with ID lower than the ack we just recieved
		for i := v.NextID; i > 0; i-- {
			// we look for an ID lower than v.NextID
			ackIdentifier := u.GetAckIdentifier(v.Identifier, addrStr, i)
			// if it is pending, we acknowledge it by writing to the channel
			g.ACKMutex.Lock()
			va, ok := g.PendingGossip[ackIdentifier]
			g.ACKMutex.Unlock()
			if ok {
				// write to the corresponding channel to stop timer in
				//rumor message
				*(va.Channel) <- true
				// set the initial message to the first message acked
				initialPacket = va.InitialMessage
				// this status message is an acknowledgement
				ack = true
			}
		}
	}

	//check if peer wants packets that g have, and send them if any
	for _, v := range status.Want {
		// if v.NextID is lower than the message we want, then we have stored
		// the message that is wanted, so we send it to peer and return
		//g.WantListMutex.Lock()
		wantedID := g.WantList[v.Identifier]
		//g.WantListMutex.Unlock()
		if v.NextID < wantedID {
			// reference of the message to recover from history

			g.HistoryMutex.Lock()
			gp := g.PacketHistory[v.Identifier][v.NextID]
			g.HistoryMutex.Unlock()
			if ack {
				g.Monger(&gp, &initialPacket, addr)
			} else { // anti entropy
				g.Monger(&gp, &gp, addr)
			}
			return
		}
	}

	// iterate over g's wantlist, and look for files that peer doesn't have,
	// and send it, in the case where peer doesn't know a packet origin

	//g.WantListMutex.Lock()
	for k, v := range g.WantList {
		if v > 1 {
			found := false
			// check if we can find the identifier in the status packet
			for _, o := range status.Want {
				if o.Identifier == k {
					found = true
					break
				}
			}
			// if identifier not found in status, send 1st packet of that identifier
			if !found {
				// load packet from history
				g.HistoryMutex.Lock()
				gp := g.PacketHistory[k][uint32(1)]
				g.HistoryMutex.Unlock()
				// mongers it to the peer
				if ack {
					g.Monger(&gp, &initialPacket, addr)
				} else { // anti entropy
					g.Monger(&gp, &gp, addr)
				}
				return
			}
		}
	}
	//g.WantListMutex.Unlock()

	// check if g is late on peer, and request messages if true
	for _, v := range status.Want {
		// if nextID > a message we want, request it
		//g.WantListMutex.Lock()
		wantedID := g.WantList[v.Identifier]
		//g.WantListMutex.Unlock()
		if v.NextID > wantedID {
			// send status to request it
			g.SendStatus(&addr)
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

		// monger initial message to random peer
		g.Monger(&initialPacket, &initialPacket, addr)
	}
}
