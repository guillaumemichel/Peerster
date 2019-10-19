package gossiper

import (
	"fmt"

	u "github.com/guillaumemichel/Peerster/utils"
)

// PrintPeers : print the known peers from the gossiper
func (g *Gossiper) PrintPeers() {
	fmt.Println("PEERS", g.PeersToString())
}

// PeersToString : return a string containing the list of known peers
func (g *Gossiper) PeersToString() string {
	str := ""
	// if not peers return empty string
	if len(g.Peers) == 0 {
		return str
	}
	for _, v := range g.Peers {
		str += v.String() + ","
	}
	// don't return the last ","
	str = str[:len(str)-1]
	return str
}

// PrintMessageClient : print messages from the client
func (g *Gossiper) PrintMessageClient(text string) {
	fmt.Println("CLIENT MESSAGE", text)
	g.PrintPeers()
}

// PrintSimpleMessage : print simple messages received from gossipers
func (g *Gossiper) PrintSimpleMessage(msg u.SimpleMessage, from string) {

	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		msg.OriginalName, from, msg.Contents)

	g.PrintPeers()
}

// PrintRumorMessage : print rumor messages received from gossipers
func (g *Gossiper) PrintRumorMessage(msg u.RumorMessage, from string) {
	fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n",
		msg.Origin, from, msg.ID, msg.Text)
	g.PrintPeers()
}

// PrintStatusMessage : print status messages received from gossipers
func (g *Gossiper) PrintStatusMessage(msg u.StatusPacket, from string) {
	fmt.Printf("STATUS from %s", from)
	for _, v := range msg.Want {
		fmt.Printf(" peer %s nextID %d", v.Identifier, v.NextID)
	}
	fmt.Println()
	g.PrintPeers()
}

// PrintFlippedCoin : prints flipped coin message
func (g *Gossiper) PrintFlippedCoin(addr string) {
	fmt.Printf("FLIPPED COIN sending rumor to %s\n", addr)
}

// PrintInSync : prints in sync message
func (g *Gossiper) PrintInSync(addr string) {
	fmt.Printf("IN SYNC WITH %s\n", addr)
}

// PrintUpdateRoute prints the DSDV update message
func (g *Gossiper) PrintUpdateRoute(origin, addr string) {
	fmt.Printf("DSDV %s %s\n", origin, addr)
}
