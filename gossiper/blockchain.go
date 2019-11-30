package gossiper

import (
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ManageTLC sends TLCmessages and wait for acks
func (g *Gossiper) ManageTLC(filename string, size int64,
	metafilehash u.ShaHash) {

	// Create the Tx block
	tx := u.TxPublish{
		Name:         filename,
		Size:         size,
		MetafileHash: metafilehash[:],
	}

	// if Hw3ex2 skip this
	if g.Hw3ex3 {
		// handles everything
		if g.TLCReady {
			// no message has been sent yet at this round

			// one message will been sent for the round
			g.TLCReady = false
			g.BuildAndSendTLC(tx)
		} else {
			g.OwnTLCBuffer.Push(&tx)
			c := make(chan bool)
			g.TLCWaitChan[&tx] = c
			// return only when ok

			//TODO test this
			select {
			case <-c:

			}
		}
	} else {
		g.BuildAndSendTLC(tx)
	}
}

// BuildAndSendTLC BuildAndSendTLC
func (g *Gossiper) BuildAndSendTLC(tx u.TxPublish) {
	// create an empty hash
	var zeroes u.ShaHash
	for i := range zeroes[:] {
		zeroes[i] = 0
	}

	// create the block to publish
	bp := u.BlockPublish{
		PrevHash:    zeroes,
		Transaction: tx,
	}

	g.WantListMutex.Lock()
	id := g.WantList[g.Name]
	g.WantList[g.Name]++
	g.WantListMutex.Unlock()

	var sp *u.StatusPacket
	if g.Hw3ex3 {
		status := g.BuildStatusPacket()
		sp = &status
	} else {
		sp = nil
	}

	// Create the TLC message
	tlc := u.TLCMessage{
		Origin:      g.Name,
		ID:          id,
		Confirmed:   u.UnconfirmedInt,
		TxBlock:     bp,
		VectorClock: sp,
		Fitness:     0,
	}

	c := make(chan u.TLCAck)
	// register channel
	g.BlockChans[id] = &c
	timeoutChan := make(chan bool)

	acks := make(map[string]bool)
	acks[g.Name] = true

	// wait for a majority of acks
	for len(acks) < g.Majority {
		if g.ShouldPrint(logHW3, 2) {
			g.Printer.Println("Sending TLC")
		}
		g.SendTLC(tlc)
		// timeout function
		go func() {
			// wait for timeout
			time.Sleep(g.StubbornTimeout)
			timeoutChan <- true
		}()
		// collect all acks
		timeout := false
		for !timeout {
			select {
			case ack := <-c:
				acks[ack.Origin] = true
				if len(acks) >= g.Majority {
					timeout = true
				}
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Got", len(acks), "acks, need",
						g.Majority)
				}
			case <-timeoutChan:
				timeout = true
			}
		}
	}
	// delete channel
	delete(g.BlockChans, id)

	// confirm tcl to all peers
	tlc.Confirmed = int(id)

	// select new message id
	g.WantListMutex.Lock()
	tlc.ID = g.WantList[g.Name]
	g.WantList[g.Name]++
	g.WantListMutex.Unlock()

	// store the confirmed message
	g.ConfirmedTLC[g.Name][g.TLCRounds[g.Name]] = tlc
	g.TLCAcksPerRound[g.TLCRounds[g.Name]]++
	g.CheckChangeRound()

	// get a list of names from a map
	names := make([]string, len(acks))
	i := 0
	for name := range acks {
		names[i] = name
		i++
	}
	// print message and broadcast confirmed message
	g.PrintReBroadcastID(int(id), names)
	g.SendTLC(tlc)
	// return to function that adds file to gossiper
}

// CheckChangeRound check if we can move a round forward
func (g *Gossiper) CheckChangeRound() {
	// if origin is at the same round as me, check if I can progress
	// to next round
	round := g.TLCRounds[g.Name]
	count := g.TLCAcksPerRound[round]
	// if I have a majority of confirmed messages from this round
	if count >= g.Majority {
		// I can progress to next round
		g.TLCReady = true
		// mytime++
		g.TLCRounds[g.Name]++

		// prepare print
		confirmed := make([]u.PeerStatus, count)
		i := 0
		for k, v := range g.ConfirmedTLC {
			if tlc, ok := v[round]; ok {
				confirmed[i] = u.PeerStatus{Identifier: k, NextID: tlc.ID}
				i++
			}
		}
		// print message
		g.PrintAdvancingToRound(round+1, confirmed)

		// send next request in the queue
		tlc := g.OwnTLCBuffer.Pop()
		if tlc != nil {
			// tell the guy stuck in ManageTLC to continue
			g.TLCWaitChan[tlc] <- true
		}
	}

}

// CheckTLCStatus CheckTLCStatus
func (g *Gossiper) CheckTLCStatus(builtOn []u.PeerStatus) bool {
	// iterate over builtOn
	for _, s := range builtOn {
		// if an element of builtOn is more recent that the last one we have
		// from that host, return false -> don't confirm
		if s.NextID > g.WantList[s.Identifier] {
			return false
		}
	}
	return true
}

// HandleTLCMessage HandleTLCMessage
func (g *Gossiper) HandleTLCMessage(gp u.GossipPacket) {
	tlc := *gp.TLCMessage
	if g.WriteGossipToHistory(gp) {
		// prints tlc message to console
		g.PrintFreshTLC(tlc)

		if g.Hw3ex3 {
			if tlc.Confirmed > 0 && tlc.Origin != g.Name {
				//confirmed message, increase the counter

				round, ok := g.TLCRounds[tlc.Origin]
				// if no message for origin create entry
				if !ok {
					round = 0
				}
				// if no message at this round create entry
				c, ok := g.TLCAcksPerRound[round]
				if !ok {
					c = 0
				}
				// increase the number of confirmed messages at this round
				g.TLCAcksPerRound[round] = c + 1
				// increase the round counter of origin
				g.TLCRounds[tlc.Origin] = round + 1

				// store the confirmed tlc
				if _, ok = g.ConfirmedTLC[tlc.Origin]; !ok {
					g.ConfirmedTLC[tlc.Origin] = make(map[int]u.TLCMessage)
				}
				g.ConfirmedTLC[tlc.Origin][round] = tlc

				// trigger majority check
			}
		}

		// ack the message
		if tlc.Confirmed < 0 {
			// careful here
			g.AckTLC(tlc)
		}
		// monger to random peer
		g.Monger(&gp, &gp, *g.GetRandPeer())
	}

}

// VerifyConfirmTLC VerifyConfirmTLC
func (g *Gossiper) VerifyConfirmTLC() {
	for tlc, round := range g.OutTLCBuffer {
		// if not a message from the past, and vector clock ok
		if round >= g.TLCRounds[g.Name] &&
			g.CheckTLCStatus(tlc.VectorClock.Want) {

			// send ack
			g.SendTLCAck(*tlc)
			// delete entry
			delete(g.OutTLCBuffer, tlc)
		} else if round < g.TLCRounds[g.Name] {
			// delete messages from the past
			delete(g.OutTLCBuffer, tlc)
		}
	}
}

// AckTLC acks a TLC message
func (g *Gossiper) AckTLC(tlc u.TLCMessage) {
	if g.Hw3ex3 && g.AckAll {
		// don't acknowledge message from the past
		if g.TLCRounds[tlc.Origin] < g.TLCRounds[g.Name] {
			return
		}

		// don't ack all messages
		if !g.CheckTLCStatus(tlc.VectorClock.Want) {
			// add it to the queue
			g.OutTLCBuffer[&tlc] = g.TLCRounds[tlc.Origin]
			return
		}

	}

	g.SendTLCAck(tlc)
}

// SendTLCAck SendTLCAck
func (g *Gossiper) SendTLCAck(tlc u.TLCMessage) {
	// create the ack
	ack := u.TLCAck{
		Origin:      g.Name,
		ID:          tlc.ID,
		Text:        "",
		Destination: tlc.Origin,
		HopLimit:    g.AckHopLimit,
	}
	gp := u.GossipPacket{Ack: &ack}
	g.PrintSendingTLCAck(tlc.Origin, int(tlc.ID))
	// route packet to its destination
	g.RoutePacket(tlc.Origin, gp)
}

// HandleTLCAcks HandleTLCAcks
func (g *Gossiper) HandleTLCAcks(ack u.TLCAck) {
	if ack.Destination == g.Name {
		// ack for me
		if g.ShouldPrint(logHW3, 2) {
			g.Printer.Println("I got an ack")
		}
		if c, ok := g.BlockChans[ack.ID]; ok {
			*c <- ack
		}
	} else {
		// forward packet
		ack.HopLimit--
		// if positive forward it, otherwise drop it
		if ack.HopLimit > 0 {
			gp := u.GossipPacket{Ack: &ack}
			g.RoutePacket(ack.Destination, gp)
		}
	}
}
