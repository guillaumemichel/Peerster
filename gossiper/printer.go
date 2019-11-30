package gossiper

import (
	"encoding/hex"
	"strconv"

	u "github.com/guillaumemichel/Peerster/utils"
)

// PrintPeers : print the known peers from the gossiper
func (g *Gossiper) PrintPeers() {
	g.Printer.Println("PEERS", g.PeersToString())
}

// PeersToString : return a string containing the list of known peers
func (g *Gossiper) PeersToString() string {
	str := ""
	// if not peers return empty string
	//g.PeerMutex.Lock()
	l := len(g.Peers)
	//g.PeerMutex.Unlock()
	if l == 0 {
		return str
	}
	//g.PeerMutex.Lock()
	for _, v := range g.Peers {
		str += v.String() + ","
	}
	//g.PeerMutex.Unlock()
	// don't return the last ","
	str = str[:len(str)-1]
	return str
}

// PrintMessageClient : print messages from the client
func (g *Gossiper) PrintMessageClient(text string) {
	if g.ShouldPrint(logHW1, 1) {
		g.Printer.Println("CLIENT MESSAGE", text)
		g.PrintPeers()
	}
}

// PrintSimpleMessage : print simple messages received from gossipers
func (g *Gossiper) PrintSimpleMessage(msg u.SimpleMessage, from string) {

	if g.ShouldPrint(logHW1, 1) {
		g.Printer.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
			msg.OriginalName, from, msg.Contents)

		g.PrintPeers()
	}
}

// PrintRumorMessage : print rumor messages received from gossipers
func (g *Gossiper) PrintRumorMessage(msg u.RumorMessage, from string) {
	if g.ShouldPrint(logHW1, 1) || g.ShouldPrint(logHW2, 1) {

		g.Printer.Printf("RUMOR origin %s from %s ID %d contents %s\n",
			msg.Origin, from, msg.ID, msg.Text)
		g.PrintPeers()
	}
}

// PrintStatusMessage : print status messages received from gossipers
func (g *Gossiper) PrintStatusMessage(msg u.StatusPacket, from string) {
	if g.ShouldPrint(logHW1, 1) {
		str := "STATUS from " + from
		for _, v := range msg.Want {
			str += " peer " + v.Identifier + " nextID " +
				strconv.Itoa(int(v.NextID))
		}
		g.Printer.Println(str)
		g.PrintPeers()
	}
}

// PrintMongering : prints mongering message
func (g *Gossiper) PrintMongering(addr string) {
	if g.ShouldPrint(logHW1, 1) {
		g.Printer.Printf("MONGERING with %s\n", addr)
	}
}

// PrintFlippedCoin : prints flipped coin message
func (g *Gossiper) PrintFlippedCoin(addr string) {
	if g.ShouldPrint(logHW1, 1) {
		g.Printer.Printf("FLIPPED COIN sending rumor to %s\n", addr)
	}
}

// PrintInSync : prints in sync message
func (g *Gossiper) PrintInSync(addr string) {
	if g.ShouldPrint(logHW1, 1) {
		g.Printer.Printf("IN SYNC WITH %s\n", addr)
	}
}

// PrintUpdateRoute prints the DSDV update message
func (g *Gossiper) PrintUpdateRoute(origin, addr string) {
	if g.ShouldPrint(logHW2, 1) {
		g.Printer.Printf("DSDV %s %s\n", origin, addr)
	}
}

// PrintPrivateMessage prints the private message to destination host
func (g *Gossiper) PrintPrivateMessage(pm u.PrivateMessage) {
	if g.ShouldPrint(logHW2, 1) {
		g.Printer.Printf("PRIVATE origin %s hop-limit %d contents %s\n",
			pm.Origin, pm.HopLimit, pm.Text)
	}
}

// PrintDownloadMetaFile prints downloading metafile
func (g *Gossiper) PrintDownloadMetaFile(dest, name string) {
	if g.ShouldPrint(logHW2, 1) || g.ShouldPrint(logHW3, 1) {
		g.Printer.Printf("DOWNLOADING metafile of %s from %s\n", name, dest)
	}
}

// PrintDownloadChunk prints downloading chunk number n message
func (g *Gossiper) PrintDownloadChunk(dest, name string, n int) {
	if g.ShouldPrint(logHW2, 1) || g.ShouldPrint(logHW3, 1) {
		g.Printer.Printf("DOWNLOADING %s chunk %d from %s\n", name, n, dest)
	}
}

// PrintReconstructFile print reconstructed file message
func (g *Gossiper) PrintReconstructFile(name string) {
	if g.ShouldPrint(logHW2, 1) || g.ShouldPrint(logHW3, 1) {
		g.Printer.Printf("RECONSTRUCTED file %s\n", name)
	}
}

// PrintUnknownMode unknown mode message
func (g *Gossiper) PrintUnknownMode() {
	g.Printer.Println("Error: Unknown gossiper mode!")
}

// PrintExpectedRumorMode print expected rumor mode message if in simple mode
// and complicated message is received
func (g *Gossiper) PrintExpectedRumorMode(message string) {
	g.Printer.Println("Warning: gossiper in simple mode received a", message,
		", discarding it")
}

// PrintSentPrivateMessage print leaving private message
func (g *Gossiper) PrintSentPrivateMessage(dest, text string) {
	if g.ShouldPrint(logHW2, 1) {
		g.Printer.Printf("CLIENT MESSAGE %s dest %s\n", text, dest)
	}
}

// PrintHashOfIndexedFile print the hash of an indexed file
func (g *Gossiper) PrintHashOfIndexedFile(file, hash string, n int) {
	if g.ShouldPrint(logHW2, 1) || g.ShouldPrint(logHW3, 1) {

		g.Printer.Printf("INDEXED file %s, hash is %s, %d chunks\n",
			file, hash, n)
	}
}

// PrintFoundMatch print found match after file search
func (g *Gossiper) PrintFoundMatch(filename, peer, metahash string,
	chunks []uint64) {
	if g.ShouldPrint(logHW3, 1) {

		c := ""
		for i, v := range chunks {
			if i != 0 {
				c += ","
			}
			c += strconv.Itoa(int(v))
		}
		g.Printer.Printf("FOUND match %s at %s metafile=%s chunks=%s\n",
			filename, peer, metahash, c)
	}
}

// PrintSearchFinished print search finished message
func (g *Gossiper) PrintSearchFinished() {
	if g.ShouldPrint(logHW3, 1) {
		g.Printer.Println("SEARCH FINISHED")
	}
}

// PrintUnconfirmedGossip print unconfirmed gossipe message
func (g *Gossiper) PrintUnconfirmedGossip(origin, filename, metahash string,
	id, size int) {
	if g.ShouldPrint(logHW3, 1) {
		g.Printer.Printf("UNCONFIRMED GOSSIP origin %s ID %d file name %s "+
			"size %d metahash %s\n", origin, id, filename, size, metahash)
	}
}

// PrintConfirmedGossip message
func (g *Gossiper) PrintConfirmedGossip(origin, filename, metahash string,
	id, size int) {
	if g.ShouldPrint(logHW3, 1) {
		g.Printer.Printf("CONFIRMED GOSSIP origin %s ID %d file name %s size"+
			"%d metahash %s\n", origin, id, filename, size, metahash)
	}
}

// PrintSendingTLCAck message
func (g *Gossiper) PrintSendingTLCAck(origin string, id int) {
	if g.ShouldPrint(logHW3, 1) {
		g.Printer.Printf("SENDING ACK origin %s ID %d\n", origin, id)
	}
}

// PrintReBroadcastID message
func (g *Gossiper) PrintReBroadcastID(id int, names []string) {
	if g.ShouldPrint(logHW3, 1) {
		list := ""
		for _, n := range names {
			list += n + ","
		}
		g.Printer.Printf("RE-BROADCAST ID %d WITNESSES %s\n",
			id, list[:len(list)-1])
	}
}

// PrintFreshTLC print TLC message on reception
func (g *Gossiper) PrintFreshTLC(tlc u.TLCMessage) {
	tx := tlc.TxBlock.Transaction

	origin := tlc.Origin
	id := tlc.ID
	if tlc.Confirmed >= 0 {
		g.PrintConfirmedGossip(origin, tx.Name,
			hex.EncodeToString(tx.MetafileHash), int(id), int(tx.Size))
	} else {
		g.PrintUnconfirmedGossip(origin, tx.Name,
			hex.EncodeToString(tx.MetafileHash), int(id), int(tx.Size))
	}
}

// NoKnownPeer message
func (g *Gossiper) NoKnownPeer() {
	if g.ShouldPrint(logHW1, 1) || g.ShouldPrint(logHW2, 1) ||
		g.ShouldPrint(logHW3, 1) {
		g.Printer.Println("No known peer to send messages")
	}
}
