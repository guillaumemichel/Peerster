package gossiper

import (
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ManageTCL sends TLCmessages and wait for acks
func (g *Gossiper) ManageTCL(filename string, size int64,
	metafilehash u.ShaHash) bool {

	// Create the Tx block
	tx := u.TxPublish{
		Name:         filename,
		Size:         size,
		MetafileHash: metafilehash[:],
	}

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
	//id := g.Round

	g.WantListMutex.Lock()
	id := g.WantList[g.Name]
	g.WantList[g.Name]++
	g.WantListMutex.Unlock()

	// Create the TLC message
	tlc := u.TLCMessage{
		Origin:      g.Name,
		ID:          id,
		Confirmed:   u.UnconfirmedInt,
		TxBlock:     bp,
		VectorClock: nil,
		Fitness:     0,
	}

	c := make(chan u.TLCAck)
	// register channel
	g.BlockChans[g.Round] = &c
	timeoutChan := make(chan bool)

	acks := make(map[string]bool)
	acks[g.Name] = true
	majority := g.N/2 + 1 // more than 50%

	// wait for a majority of acks
	for len(acks) < majority {
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
			case <-timeoutChan:
				timeout = true
			}
		}
	}
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
	g.SendTLC(tlc)
	// return to function that adds file to gossiper
	return true
}

// AckTLC acks a TLC message
func (g *Gossiper) AckTLC(tlc u.TLCMessage) {
	// create the ack
	ack := u.TLCAck{
		Origin:      g.Name,
		ID:          tlc.ID,
		Text:        "",
		Destination: tlc.Origin,
		HopLimit:    u.DefaultHopLimit,
	}
	gp := u.GossipPacket{Ack: &ack}
	// route packet to its destination
	g.RoutePacket(tlc.Origin, gp)
}
