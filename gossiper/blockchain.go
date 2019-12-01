package gossiper

import (
	"math/rand"
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

	var hash u.ShaHash
	if g.GotConsensus {
		// take last block of committed history
		hash = g.CommittedHistory.Start.Hash
	} else {
		// build on best block of stage s+1
		hash = g.SelectedTLC.TxBlock.Hash()
	}

	// create the block to publish
	bp := u.BlockPublish{
		// zeros if history not filled
		PrevHash:    hash,
		Transaction: tx,
	}

	// if Hw3ex2 skip this
	if g.Hw3ex3 {
		// handles everything
		if g.TLCReady {
			// no message has been sent yet at this round

			// one message will been sent for the round
			g.TLCReady = false
		} else {
			if g.ShouldPrint(logHW3, 2) {
				g.Printer.Println("Adding index request to ownbuffer")
			}
			g.OwnTLCBuffer.Push(&tx)
			c := make(chan bool)
			g.TLCWaitChan[&tx] = c
			// return only when ok

			// kick current search
			if g.TLCAcksPerRound[g.TLCRounds[g.Name]] >= g.Majority {
				if len(g.BlockChans) == 1 {
					for _, v := range g.BlockChans {
						if g.ShouldPrint(logHW3, 2) {
							g.Printer.Println("Sending abort signal to " +
								"current unconfirmed TLC")
						}
						dummyAck := u.TLCAck{}
						*v[1] <- dummyAck
					}
				} else {
					g.Printer.Println("Wrong size of BlockChans (>1)")
				}
			}

			//TODO test this
			select {
			case <-c:
				g.TLCReady = false
				delete(g.TLCWaitChan, &tx)
			}
		}
	}
	g.BuildAndSendTLC(bp, nil, u.UnconfirmedInt)

}

// BuildAndSendTLC BuildAndSendTLC
func (g *Gossiper) BuildAndSendTLC(bp u.BlockPublish, fit *float32,
	confirmed int) {

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

	if fit == nil {
		var n float32
		if g.Hw3ex4 {
			// if ex4, get random fitness value
			n = rand.Float32()
		}
		fit = &n
	}

	// Create the TLC message
	tlc := u.TLCMessage{
		Origin:      g.Name,
		ID:          id,
		Confirmed:   confirmed,
		TxBlock:     bp,
		VectorClock: sp,
		Fitness:     *fit,
	}

	cAck := make(chan u.TLCAck)
	cAbort := make(chan u.TLCAck)

	// register channel
	g.BlockChans[id] = make([]*chan u.TLCAck, 2)
	g.BlockChans[id][0] = &cAck
	g.BlockChans[id][1] = &cAbort
	timeoutChan := make(chan bool)

	acks := make(map[string]bool)
	acks[g.Name] = true

	broadcast := true

	// wait for a majority of acks
	for len(acks) < g.Majority && broadcast {
		if g.ShouldPrint(logHW3, 3) {
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
			if g.ShouldPrint(logHW3, 3) {
				g.Printer.Println("No Timeout")
			}
			select {
			case ack := <-cAck:
				acks[ack.Origin] = true
				if len(acks) >= g.Majority {
					timeout = true
				}
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Got", len(acks), "acks, need",
						g.Majority)
				}
			case <-cAbort:
				if g.ShouldPrint(logHW3, 3) {
					g.Printer.Println("Aborting TLC")
				}
				// don't broadcast
				broadcast = false
				// get out of here
				timeout = true
			case <-timeoutChan:
				if g.ShouldPrint(logHW3, 3) {
					g.Printer.Println("Timeout")
				}
				timeout = true
			}
		}
	}
	if g.ShouldPrint(logHW3, 3) {
		g.Printer.Println("Done TLC")
	}
	// delete channel
	delete(g.BlockChans, id)

	if broadcast {
		// confirm tcl to all peers
		tlc.Confirmed = int(id)

		// select new message id
		g.WantListMutex.Lock()
		tlc.ID = g.WantList[g.Name]
		g.WantList[g.Name]++
		g.WantListMutex.Unlock()

		// get a list of names from a map
		names := make([]string, len(acks))
		i := 0
		for name := range acks {
			names[i] = name
			i++
		}
		// print message and broadcast confirmed message
		g.PrintReBroadcastID(int(id), names)

		// store the confirmed message
		g.ConfirmedTLC[g.Name][g.TLCRounds[g.Name]] = tlc
		g.TLCAcksPerRound[g.TLCRounds[g.Name]]++
		g.CheckChangeRound()

		g.SendTLC(tlc)
		// return to function that adds file to gossiper
	} else {
		g.CheckChangeRound()

	}
}

// CheckChangeRound check if we can move a round forward
func (g *Gossiper) CheckChangeRound() {
	// if origin is at the same round as me, check if I can progress
	// to next round
	round := g.TLCRounds[g.Name]
	count := g.TLCAcksPerRound[round]
	// if I have a majority of confirmed messages from this round
	// if I have already sent a message at this round
	if count >= g.Majority && !g.TLCReady {
		// mytime++
		g.TLCRounds[g.Name]++

		// prepare print
		confirmed := make([]u.TLCMessage, count)
		i := 0
		for _, v := range g.ConfirmedTLC {
			if tlc, ok := v[round]; ok {
				confirmed[i] = tlc
				i++
			}
		}
		if g.Hw3ex4 {
			if g.QSCStage != 2 {
				// get message with highest fitness
				var selected u.TLCMessage
				var max float32 = -1.0
				for _, v := range confirmed {
					if v.Fitness > max {
						max = v.Fitness
						selected = v
					}
				}
				// s -> s+1 or s+1 -> s+2
				g.QSCStage++

				g.SelectedTLC = selected

				// print message
				g.PrintAdvancingToRound(round+1, confirmed)

				// send TLC confirming message with highest fitness
				g.BuildAndSendTLC(selected.TxBlock, &selected.Fitness,
					selected.Confirmed)
				return
			}

			// select message (concensus)
			agree := make(map[string]bool)
			// set of origin that agree upon g.SelectedTLC
			selectedHash := g.SelectedTLC.TxBlock.Hash()
			for _, v := range g.ConfirmedTLC {
				for i := 0; i < 2; i++ {
					// check how many nodes share this confirmed message in the
					// last 2 rounds
					if tlc, ok := v[round-i]; ok {
						if tlc.TxBlock.Hash() == selectedHash {
							agree[tlc.Origin] = true
						}
					}
				}
			}
			// consensus
			if len(agree) >= g.Majority {

				currBlock := g.SelectedTLC.TxBlock
				// iterate in reverse order to write to committed history
				blockList := []u.BlockPublish{currBlock}
				// while the block do not refer to the last one of history
				security := 0
				for currBlock.PrevHash != g.CommittedHistory.Tail.Hash &&
					security < u.SecurityLimit {
					p, ok := g.HashToBlock[currBlock.PrevHash]
					if !ok {
						g.Printer.Println("Fatal error: didn't found block I" +
							"have to write to history")
						break
					}
					currBlock = *p
					// append block to list to write
					blockList = append(blockList, currBlock)
					security++
				}

				// write all blocks to committed history in reverse order
				for i := len(blockList) - 1; i >= 0; i-- {
					// write block to committed history
					newBlock := u.Node{
						Block: blockList[i],
						Hash:  blockList[i].Hash(),
					}
					g.CommittedHistory.InsertNode(&newBlock)
				}

				g.GotConsensus = true

				// print concensus message
				g.PrintConcensus(g.SelectedTLC, round,
					g.CommittedHistory.Filenames)
			} else {
				g.GotConsensus = false
			}
			// s+2 -> s next QSC round
			g.QSCStage = 0

		}
		// print message
		g.PrintAdvancingToRound(round+1, confirmed)

		g.NextTLC()

	}

}

// NextTLC send next TLC message from the queue if any
func (g *Gossiper) NextTLC() {
	// I can progress to next round
	g.TLCReady = true

	// send next request in the queue
	tlc := g.OwnTLCBuffer.Pop()
	if tlc != nil {
		// tell the guy stuck in ManageTLC to continue
		g.TLCWaitChan[tlc] <- true
	}
}

// CheckTLCStatus CheckTLCStatus
func (g *Gossiper) CheckTLCStatus(builtOn []u.PeerStatus) bool {
	// iterate over builtOn
	for _, s := range builtOn {
		// if an element of builtOn is more recent that the last one we have
		// from that host, return false -> don't confirm
		g.WantListMutex.Lock()
		id := g.WantList[s.Identifier]
		g.WantListMutex.Unlock()
		if s.NextID > id {
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

				// to be able to recover a block from its hash
				g.HashToBlock[tlc.TxBlock.Hash()] = &tlc.TxBlock

				round, ok := g.TLCRounds[tlc.Origin]
				// if no message for origin create entry
				if !ok {
					round = 0
				}

				already := false
				if v, ok := g.PacketHistory[tlc.Origin][uint32(
					tlc.Confirmed)]; ok && v.TLCMessage != nil &&
					int(v.TLCMessage.ID) == tlc.Confirmed {
					// decrease round, because message already received as not
					// confirmed
					round--
					already = true
				}

				// if no message at this round create entry
				c, ok := g.TLCAcksPerRound[round]
				if !ok {
					c = 0
				}

				// increase the number of confirmed messages at this round
				g.TLCAcksPerRound[round] = c + 1

				if !already {
					// increase the round counter of origin
					g.TLCRounds[tlc.Origin] = round + 1
				}

				// store the confirmed tlc
				if _, ok = g.ConfirmedTLC[tlc.Origin]; !ok {
					g.ConfirmedTLC[tlc.Origin] = make(map[int]u.TLCMessage)
				}
				g.ConfirmedTLC[tlc.Origin][round] = tlc

				// trigger majority check
				g.CheckChangeRound()
			}
		}

		// ack the message
		if tlc.Confirmed < 0 {
			//unconfirmed tlc

			// careful with rounds
			g.AckTLC(tlc)
			if g.Hw3ex3 {
				found := false
				// init TLCRounds of origin if not initialized yet
				if _, ok := g.TLCRounds[tlc.Origin]; !ok {
					g.TLCRounds[tlc.Origin] = 0
				}

				// check if g already got the confirmed version of tlc
				for _, v := range g.ConfirmedTLC[tlc.Origin] {
					if int(tlc.ID) == v.Confirmed {
						// already counted in the round history of origin
						found = true
					}
				}
				// we don't have the confirmation of tlc yet
				if !found {
					// increase round of origin
					g.TLCRounds[tlc.Origin]++
				}
			}
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
	if g.Hw3ex3 && !g.AckAll {
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
	if g.Hw3ex4 {
		for _, name := range g.CommittedHistory.Filenames {
			if tlc.TxBlock.Transaction.Name == name {
				// name already registered, don't ack
				return
			}
		}
		currBlock := tlc.TxBlock
		security := 0
		for currBlock.Hash() != g.CommittedHistory.Tail.Hash &&
			security < u.SecurityLimit {
			p, ok := g.HashToBlock[currBlock.PrevHash]
			// hash not known
			if !ok {
				// don't ack that shit
				return
			}
			currBlock = *p
			security++
		}
		if security == u.SecurityLimit {
			// may be a loop or too long chain
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
			*c[0] <- ack
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
