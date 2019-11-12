package gossiper

import (
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
	g.Printer.Println("CLIENT MESSAGE", text)
	g.PrintPeers()
}

// PrintSimpleMessage : print simple messages received from gossipers
func (g *Gossiper) PrintSimpleMessage(msg u.SimpleMessage, from string) {

	g.Printer.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		msg.OriginalName, from, msg.Contents)

	g.PrintPeers()
}

// PrintRumorMessage : print rumor messages received from gossipers
func (g *Gossiper) PrintRumorMessage(msg u.RumorMessage, from string) {
	g.Printer.Printf("RUMOR origin %s from %s ID %d contents %s\n",
		msg.Origin, from, msg.ID, msg.Text)
	g.PrintPeers()
}

// PrintStatusMessage : print status messages received from gossipers
func (g *Gossiper) PrintStatusMessage(msg u.StatusPacket, from string) {
	str := "STATUS from " + from
	for _, v := range msg.Want {
		str += " peer " + v.Identifier + " nextID " +
			strconv.Itoa(int(v.NextID))
	}
	g.Printer.Println(str)
	g.PrintPeers()
}

// PrintFlippedCoin : prints flipped coin message
func (g *Gossiper) PrintFlippedCoin(addr string) {
	g.Printer.Printf("FLIPPED COIN sending rumor to %s\n", addr)
}

// PrintInSync : prints in sync message
func (g *Gossiper) PrintInSync(addr string) {
	g.Printer.Printf("IN SYNC WITH %s\n", addr)
}

// PrintUpdateRoute prints the DSDV update message
func (g *Gossiper) PrintUpdateRoute(origin, addr string) {
	g.Printer.Printf("DSDV %s %s\n", origin, addr)
}

// PrintPrivateMessage prints the private message to destination host
func (g *Gossiper) PrintPrivateMessage(pm u.PrivateMessage) {
	g.Printer.Printf("PRIVATE origin %s hop-limit %d contents %s\n",
		pm.Origin, pm.HopLimit, pm.Text)
}

// PrintDownloadMetaFile prints downloading metafile
func (g *Gossiper) PrintDownloadMetaFile(dest, name string) {
	g.Printer.Printf("DOWNLOADING metafile of %s from %s\n", name, dest)
}

// PrintDownloadChunk prints downloading chunk number n message
func (g *Gossiper) PrintDownloadChunk(dest, name string, n int) {
	g.Printer.Printf("DOWNLOADING %s chunk %d from %s\n", name, n, dest)
}

// PrintReconstructFile print reconstructed file message
func (g *Gossiper) PrintReconstructFile(name string) {
	g.Printer.Printf("RECONSTRUCTED file %s\n", name)
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
	g.Printer.Printf("CLIENT MESSAGE %s dest %s\n", text, dest)
}

// PrintHashOfIndexedFile print the hash of an indexed file
func (g *Gossiper) PrintHashOfIndexedFile(file, hash string) {
	g.Printer.Printf("INDEXED file %s, hash is %s\n", file, hash)
}
