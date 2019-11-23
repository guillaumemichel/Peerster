package gossiper

import (
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ManageTCL sends TLCmessages and wait for acks
func (g *Gossiper) ManageTCL(filename string, size int64,
	metafilehash u.ShaHash) {

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

	// Create the TLC message
	tlc := u.TLCMessage{
		Origin:      g.Name,
		ID:          g.Round,
		Confirmed:   false,
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
		for {
			select {
			case ack := <-c:
				acks[ack.Origin] = true
				continue
			case <-timeoutChan:
				break // timeout
			}
		}
	}
	// confirm tcl to all peers
	tlc.Confirmed = true
	g.SendTLC(tlc)
	// return to function that adds file to gossiper
}
