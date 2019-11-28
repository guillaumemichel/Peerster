package gossiper

import (
	"encoding/hex"
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
	id := g.Round

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

	firstTime := true

	// wait for a majority of acks
	for len(acks) < majority {
		if firstTime {
			metahash := hex.EncodeToString(metafilehash[:])
			g.PrintUnconfirmedGossip(g.Name, filename, metahash,
				int(id), int(size))
			firstTime = false
		} else {
			names := make([]string, len(acks))
			c := 0
			for n := range acks {
				names[c] = n
				c++
			}
			g.PrintReBroadcastID(int(id), names)
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
			case <-timeoutChan:
				timeout = true
			}
		}
	}
	// confirm tcl to all peers
	// WTFFFFFFFFFFFFFF ???
	tlc.Confirmed = int(id)

	g.SendTLC(tlc)
	// return to function that adds file to gossiper
	return true
}
