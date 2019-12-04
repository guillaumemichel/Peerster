package gossiper

import (
	"encoding/hex"
	"math"
	"math/rand"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ManageTLC sends TLCmessages and wait for acks
func (g *Gossiper) ManageTLC(filename string, size int64,
	metafilehash u.ShaHash) bool {

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
	if g.Hw3ex3 || g.Hw3ex4 {

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
	return g.BuildAndSendTLC(bp, nil)
}

// BuildAndSendTLC BuildAndSendTLC
func (g *Gossiper) BuildAndSendTLC(bp u.BlockPublish, fit *float32) bool {

	if g.Hw3ex4 {
		// check if name already in blockchain
		node := g.CommittedHistory.Start
		for node.Next != nil {
			if node.Block.Transaction.Name == bp.Transaction.Name {
				// don't index a file if name already in blockchain
				if g.ShouldPrint(logHW3, 1) {
					g.Printer.Println("Name already taken, aborting " +
						"indexing")
				}
				return false
			}
			node = node.Next
		}
	}

	g.WantListMutex.Lock()
	id := g.WantList[g.Name]
	g.WantList[g.Name]++
	g.WantListMutex.Unlock()

	var sp *u.StatusPacket
	if g.Hw3ex3 || g.Hw3ex4 {
		status := g.BuildStatusPacket()
		sp = &status
	} else {
		sp = nil
	}

	if fit == nil {
		var n float32
		if g.Hw3ex4 {
			rand.Seed(int64(u.GetRealRand(math.MaxInt32)))
			// if ex4, get random fitness value
			n = rand.Float32()
		}
		fit = &n
	}

	// Create the TLC message
	tlc := u.TLCMessage{
		Origin:      g.Name,
		ID:          id,
		Confirmed:   u.UnconfirmedInt,
		TxBlock:     bp,
		VectorClock: sp,
		Fitness:     *fit,
	}

	if true {
		broadcast := true

		cAck := make(chan u.TLCAck)
		cAbort := make(chan u.TLCAck)

		// register channel
		g.BlockChans[id] = make([]*chan u.TLCAck, 2)
		g.BlockChans[id][0] = &cAck
		g.BlockChans[id][1] = &cAbort
		timeoutChan := make(chan bool)

		acks := make(map[string]bool)
		acks[g.Name] = true

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

			if g.Hw3ex3 || g.Hw3ex4 {
				status := g.BuildStatusPacket()
				sp = &status
			}
			tlc.VectorClock = sp

			// get a list of names from a map
			names := make([]string, len(acks))
			i := 0
			for name := range acks {
				names[i] = name
				i++
			}
			// print message and broadcast confirmed message
			g.PrintReBroadcastID(int(id), names)
			g.PrintConfirmedGossip(tlc.Origin, bp.Transaction.Name,
				hex.EncodeToString(bp.Transaction.MetafileHash), int(tlc.ID),
				int(bp.Transaction.Size))

			// store the confirmed message
			g.ConfirmedTLC[g.Name][g.TLCRounds[g.Name]] = tlc
			if g.ShouldPrint(logHW3, 2) {
				round := g.TLCRounds[g.Name]
				// prepare print
				confirmed := make([]u.TLCMessage, 0)
				//i := 0
				for _, v := range g.ConfirmedTLC {
					if tlc, ok := v[round]; ok {
						confirmed = append(confirmed, tlc)
						//i++
					}
				}
				g.Printer.Println("Just added:", tlc)
				g.Printer.Println("Confirmed at round", round, ":",
					confirmed)
				//g.Printer.Println(g.ConfirmedTLC)
			}

			g.TLCAcksPerRound[g.TLCRounds[g.Name]]++
			g.SendTLC(tlc)

			// return to function that adds file to gossiper
		}
		if g.Hw3ex4 && broadcast && g.QSCStage == 0 {
			g.CheckChangeRound()
			if !g.QSCWaiting {
				g.QSCWaiting = true
				select {
				case hash := <-g.QSCChan:
					g.QSCWaiting = false
					if hash == bp.Hash() {
						if g.ShouldPrint(logHW3, 1) {
							g.Printer.Println("QSC round won, indexed file")
						}
						g.CheckChangeRound()
						return true
					}
					if g.ShouldPrint(logHW3, 1) {
						g.Printer.Println("QSC round lost, did not index " +
							"anything")
					}
					g.CheckChangeRound()
					return false

				}
			}
		}
		if broadcast {
			g.CheckChangeRound()
			return true
		}
		g.Printer.Println("\nNO BROADCAST")
		//g.TLCRounds[g.Name]++
		g.CheckChangeRound()

		//g.NextTLC()
		return false
	}
	return true
}

// store the confirmed message
/*
		if _, ok := g.ConfirmedTLC[tlc.Origin][g.TLCRounds[g.Name]]; ok &&
			g.ShouldPrint(logHW3, 2) {
			g.Printer.Println("Already written 3 !!!!")
		}
	g.TLCAcksPerRound[g.TLCRounds[g.Name]]++
	g.ConfirmedTLC[g.Name][g.TLCRounds[g.Name]] = tlc
	if g.ShouldPrint(logHW3, 2) {
		round := g.TLCRounds[g.Name]
		// prepare print
		confirmed := make([]u.TLCMessage, 0)
		//i := 0
		for _, v := range g.ConfirmedTLC {
			if tlc, ok := v[round]; ok {
				confirmed = append(confirmed, tlc)
				//i++
			}
		}
		g.Printer.Println("Just added:", tlc)
		g.Printer.Println("Confirmed at round", round, ":",
			confirmed)
		//g.Printer.Println(g.ConfirmedTLC)
	}

	g.PrintConfirmedGossip(tlc.Origin, tlc.TxBlock.Transaction.Name,
		hex.EncodeToString(tlc.TxBlock.Transaction.MetafileHash),
		int(tlc.ID), int(tlc.TxBlock.Transaction.Size))

	g.CheckChangeRound()
	g.SendTLC(tlc)
	return true
*/

// CheckChangeRound check if we can move a round forward
func (g *Gossiper) CheckChangeRound() {
	// if origin is at the same round as me, check if I can progress
	// to next round
	round := g.TLCRounds[g.Name]
	// prepare print
	confirmed := make([]u.TLCMessage, 0)
	//i := 0
	for _, v := range g.ConfirmedTLC {
		if tlc, ok := v[round]; ok {
			confirmed = append(confirmed, tlc)
			//i++
		}
	}
	count := len(confirmed)

	if g.ShouldPrint(logHW3, 3) {
		g.Printer.Println(count, "acks at my round", round)
	}

	// if I have a majority of confirmed messages from this round
	// if I have already sent a message at this round
	if count >= g.Majority && !g.TLCReady {
		// mytime++
		g.TLCRounds[g.Name]++

		if g.Hw3ex4 {
			if g.ShouldPrint(logHW3, 3) {
				g.Printer.Println("Entering ex4 part")
			}

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
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Sending next TLC:",
						selected.TxBlock.Transaction.Name, selected.Fitness)
				}

				// send TLC confirming message with highest fitness
				g.BuildAndSendTLC(selected.TxBlock, &selected.Fitness)
				return
			}

			// select message (concensus)
			agree := make(map[string]bool)
			// set of origin that agree upon g.SelectedTLC
			selectedHash := g.SelectedTLC.TxBlock.Hash()
			for _, v := range g.ConfirmedTLC {
				for i := 0; i <= 2; i++ {
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

				h := g.SelectedTLC.TxBlock.Hash()
				original := g.SelectedTLC
				for _, r := range g.PacketHistory {
					for _, p := range r {
						if p.TLCMessage != nil && p.TLCMessage.Confirmed > 0 &&
							p.TLCMessage.TxBlock.Hash() == h &&
							p.TLCMessage.ID < original.ID {
							original = *p.TLCMessage
						}
					}
				}

				// print concensus message
				g.PrintConcensus(original, round,
					g.CommittedHistory.Filenames)
			} else {
				if g.ShouldPrint(logHW3, 1) {
					g.Printer.Println("No Consensus =(")
				}
				g.GotConsensus = false
			}
			g.QSCChan <- g.SelectedTLC.TxBlock.Hash()
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
func (g *Gossiper) HandleTLCMessage(tlc u.TLCMessage) {
	// prints tlc message to console
	g.PrintFreshTLC(tlc)

	if g.Hw3ex3 || g.Hw3ex4 {
		if tlc.Confirmed > 0 && tlc.Origin != g.Name {
			//confirmed message, increase the counter
			if g.ShouldPrint(logHW3, 2) {
				g.Printer.Println("ACK_ID:", tlc.Confirmed)
			}

			// to be able to recover a block from its hash
			g.HashToBlock[tlc.TxBlock.Hash()] = &tlc.TxBlock

			round, ok := g.TLCRounds[tlc.Origin]
			// if no message for origin create entry
			if !ok {
				round = 0
			}

			already := false

			/*
				specialCond := true
				if g.Hw3ex4 && g.TLCRounds[tlc.Origin]%3 == 1 &&
					g.ConfirmedTLC[tlc.Origin][g.TLCRounds[tlc.Origin]-1].
						Confirmed == tlc.Confirmed {
					specialCond = false
				}
			*/

			v, ok := g.PacketHistory[tlc.Origin][uint32(tlc.Confirmed)]

			/*
				for _, o := range g.ConfirmedTLC {
					for _, v := range o {
						if v.
					}
				}
			*/

			//specialCond := g.Hw3ex4 && ok

			if ok && v.TLCMessage != nil && int(v.TLCMessage.ID) ==
				tlc.Confirmed {
				// decrease round, because message already received as not
				// confirmed
				g.Printer.Println("already")
				round--
				already = true
			}

			// if no message at this round create entry
			_, ok = g.TLCAcksPerRound[round]
			if !ok {
				g.TLCAcksPerRound[round] = 0
			}

			// increase the number of confirmed messages at this round
			g.TLCAcksPerRound[round]++

			if !already {
				// increase the round counter of origin
				g.TLCRounds[tlc.Origin]++
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println(tlc.Origin, "now at round",
						g.TLCRounds[tlc.Origin])
				}
			}

			// store the confirmed tlc
			if _, ok = g.ConfirmedTLC[tlc.Origin]; !ok {
				g.ConfirmedTLC[tlc.Origin] = make(map[int]u.TLCMessage)
			}
			/*
				if _, ok := g.ConfirmedTLC[tlc.Origin][round]; ok &&
					g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Already written 1 !!!!")
					g.Printer.Println(g.ConfirmedTLC)
				}
			*/
			g.ConfirmedTLC[tlc.Origin][round] = tlc

			if g.ShouldPrint(logHW3, 2) {
				// prepare print
				confirmed := make([]u.TLCMessage, 0)
				//i := 0
				for _, v := range g.ConfirmedTLC {
					if tlc, ok := v[round]; ok {
						confirmed = append(confirmed, tlc)
						//i++
					}
				}
				g.Printer.Println("Just added:", tlc)
				g.Printer.Println("Confirmed at round", round, ":",
					confirmed)
				//g.Printer.Println(g.ConfirmedTLC)
			}
		}
	}

	// ack the message
	if tlc.Confirmed < 0 {
		//unconfirmed tlc

		// careful with rounds
		if g.Hw3ex3 || g.Hw3ex4 {
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
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println(tlc.Origin, "now at round",
						g.TLCRounds[tlc.Origin])
				}

			}
		}
	}
	// monger to random peer
	g.CheckChangeRound()
	/*else {
		// if we already have it, confirmed at s+1 or s+2 increase number
		if g.Hw3ex4 && tlc.Confirmed > 0 && g.QSCStage != 0 {
			/*
				if _, ok := g.ConfirmedTLC[tlc.Origin][g.TLCRounds[tlc.Origin]]; ok &&
					g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Already written 2 !!!!")
				}

			g.TLCAcksPerRound[g.TLCRounds[g.Name]]++
			g.ConfirmedTLC[tlc.Origin][g.TLCRounds[tlc.Origin]] = tlc
			if g.ShouldPrint(logHW3, 2) {
				round := g.TLCRounds[tlc.Origin]
				// prepare print
				confirmed := make([]u.TLCMessage, 0)
				//i := 0
				for _, v := range g.ConfirmedTLC {
					if tlc, ok := v[round]; ok {
						confirmed = append(confirmed, tlc)
						//i++
					}
				}
				g.Printer.Println("Just added:", tlc)
				g.Printer.Println("Confirmed at round", round, ":",
					confirmed)
				//g.Printer.Println(g.ConfirmedTLC)
			}
		}
	}
	*/
	// trigger majority check
}

// VerifyConfirmTLC VerifyConfirmTLC
func (g *Gossiper) VerifyConfirmTLC() {
	for tlc, round := range g.OutTLCBuffer {
		// if not a message from the past, and vector clock ok
		if round >= g.TLCRounds[g.Name] &&
			g.CheckTLCStatus(tlc.VectorClock.Want) {

			// send ack
			g.SendTLCAck(*tlc)
			g.HandleTLCMessage(*tlc)
			// delete entry
			delete(g.OutTLCBuffer, tlc)
		} else if round < g.TLCRounds[g.Name] {
			// delete messages from the past
			delete(g.OutTLCBuffer, tlc)
		}
	}
}

// AckTLC acks a TLC message
func (g *Gossiper) AckTLC(gp u.GossipPacket) {
	tlc := *gp.TLCMessage
	if (g.Hw3ex3 && !g.AckAll) || g.Hw3ex4 {
		// don't acknowledge message from the past
		if g.TLCRounds[tlc.Origin] < g.TLCRounds[g.Name] {
			if g.ShouldPrint(logHW3, 2) {
				g.Printer.Println("Don't ack TLC: from the past")
			}
			g.HandleTLCMessage(tlc)
			return
		}

		// don't ack all messages
		if !g.CheckTLCStatus(tlc.VectorClock.Want) {
			// add it to the queue
			g.OutTLCBuffer[&tlc] = g.TLCRounds[tlc.Origin]
			if g.ShouldPrint(logHW3, 2) {
				g.Printer.Println("Don't ack TLC: vector clock not " +
					"satisfied yet")
			}
			return
		}

	}
	if tlc.Confirmed > 0 {
		g.HandleTLCMessage(tlc)
		return
	}

	if g.Hw3ex4 {
		for _, name := range g.CommittedHistory.Filenames {
			if tlc.TxBlock.Transaction.Name == name {
				// name already registered, don't ack
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Don't ack TLC: name already taken")
				}
				g.HandleTLCMessage(tlc)
				return
			}
		}
		currBlock := tlc.TxBlock
		security := 0
		for (currBlock.PrevHash != g.CommittedHistory.Tail.Hash ||
			currBlock.PrevHash != g.CommittedHistory.Start.Hash) &&
			security < u.SecurityLimit {
			p, ok := g.HashToBlock[currBlock.PrevHash]
			// hash not known
			if !ok {
				// don't ack that shit
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Don't ack TLC: unknown history")
				}
				g.HandleTLCMessage(tlc)
				return
			}
			currBlock = *p
			security++
		}
		if security == u.SecurityLimit {
			// may be a loop or too long chain
			if g.ShouldPrint(logHW3, 2) {
				g.Printer.Println("Don't ack TLC: for security reasons")
			}
			g.HandleTLCMessage(tlc)
			return
		}

	}
	g.HandleTLCMessage(tlc)
	if tlc.Confirmed < 0 {
		g.SendTLCAck(tlc)
	}
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
