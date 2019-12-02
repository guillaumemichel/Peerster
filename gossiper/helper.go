package gossiper

import (
	"log"
	"net"
	"strings"

	u "github.com/guillaumemichel/Peerster/utils"
)

// GetRandPeer : get a random peer known to g
func (g *Gossiper) GetRandPeer() *net.UDPAddr {
	//g.PeerMutex.Lock()
	l := len(g.Peers)
	//g.PeerMutex.Unlock()
	if l == 0 {
		//fmt.Println("Error: cannot send message to random peer,",
		//	"no known peer :(")
		return nil
	}
	r := u.GetRealRand(l) // GetRand(n) also possible
	//g.PeerMutex.Lock()
	target := (g.Peers)[r]
	//g.PeerMutex.Unlock()
	return &target
}

// GetNewPeers : return a list of string all new peer addresses
func (g *Gossiper) GetNewPeers(c int) []string {
	//g.PeerMutex.Lock()
	peers := g.Peers
	if c >= len(peers) {
		//g.PeerMutex.Unlock()
		return nil
	}
	var list []string
	for _, v := range peers[c:] {
		list = append(list, v.String())
	}
	//g.PeerMutex.Unlock()
	return list
}

// GetLastPeers returns a list of peers that arrived after 'last'
func (g *Gossiper) GetLastPeers(last string) []string {
	// check if last is the last peer of g
	//g.PeerMutex.Lock()
	peers := g.Peers
	if peers[len(peers)-1].String() != last {
		var addrs []string
		index := len(peers)
		for i, v := range peers {
			if i < index && v.String() == last {
				index = i
			} else if i > index {
				// write the addresses that are after the last known one
				addrs = append(addrs, v.String())
			}
		}
		//g.PeerMutex.Unlock()
		return addrs
	}
	//g.PeerMutex.Unlock()
	return nil
}

// GetDestinations return the last destinations not sync yet
func (g *Gossiper) GetDestinations(c uint32) []string {
	g.RouteMutex.Lock()
	if int(c) >= len(g.Routes) {
		g.RouteMutex.Unlock()
		return nil
	}
	dests := make([]string, len(g.Routes))
	i := 0
	for v := range g.Routes {
		dests[i] = v
		i++
	}
	g.RouteMutex.Unlock()
	return dests
}

// GetPrivateMessage returns the pms with index higher than the parameter
func (g *Gossiper) GetPrivateMessage(c int) []u.PrivateMessage {
	prvtmsg := g.PrivateMsg
	if c >= len(prvtmsg) {
		return nil
	}
	return prvtmsg[c:]
}

// AddPeer : adds the given peer to peers list if not already in it
func (g *Gossiper) AddPeer(addr *net.UDPAddr) {
	addrStr := addr.String()
	if !strings.Contains(g.PeersToString(), addrStr) &&
		addrStr != g.GossipAddr.String() && addrStr != g.GossipAddr.String() {
		//g.PeerMutex.Lock()
		g.Peers = append(g.Peers, *addr)
		//g.PeerMutex.Unlock()
	}
}

// ReceiveOK : return false if message too long to be handled, true otherwise
func (g *Gossiper) ReceiveOK(ok bool, rcvBytes []byte) bool {
	if !ok && len(rcvBytes) >= u.BufferSize {
		log.Printf(`Warning: incoming message possibly larger than %d bytes
			couldn't be read!\n`, u.BufferSize)
		return false
	}
	return true
}

/*
// RecoverHistoryRumor : recover a message from rumor history
func (g *Gossiper) RecoverHistoryRumor(ref u.MessageReference) u.RumorMessage {

	//g.HistoryMutex.Lock()
	oH, ok := g.RumorHistory.Load(ref.Origin)
	//g.HistoryMutex.Unlock()
	// if ID ref is larger than array size or message not in array
	if !ok || len(oH.([]u.HistoryMessage)) < int(ref.ID) ||
		oH.([]u.HistoryMessage)[ref.ID-1].ID != ref.ID {

		//log.Println("Error: message queried and not found in history!")
		//log.Println(ref.Origin, " : ", ref.ID, " : ",
		//	len(oH.([]u.HistoryMessage)))
		return u.RumorMessage{}
	}

	// recover the text in history to create the rumor
	rumor := u.RumorMessage{
		Origin: ref.Origin,
		ID:     ref.ID,
		Text:   oH.([]u.HistoryMessage)[ref.ID-1].Text,
	}
	return rumor
}
*/

/*
// HistoryMessageToByte : recover a rumor message from history, protobuf it
// and sends back an array of bytes
func (g *Gossiper) HistoryMessageToByte(ref u.MessageReference) []byte {
	rumor := g.RecoverHistoryRumor(ref)
	// protobuf the rumor to get a byte array
	packet := u.ProtobufGossip(&u.GossipPacket{Rumor: &rumor})
	return packet
}
*/

// WriteGossipToHistory write a gossip packet to history
func (g *Gossiper) WriteGossipToHistory(gp u.GossipPacket) bool {
	var origin string
	var id uint32

	// set origin and id according to the gossip type
	if gp.Rumor != nil {
		origin = gp.Rumor.Origin
		id = gp.Rumor.ID
	} else if gp.TLCMessage != nil {
		origin = gp.TLCMessage.Origin
		id = gp.TLCMessage.ID
	} else {
		if g.ShouldPrint(logHW3, 1) {
			g.Printer.Println("Cannot write the gossip to history invalid type")
		}
		return false
	}

	// discard messages without origin
	if origin == "" {
		return false
	}

	g.HistoryMutex.Lock()
	if _, ok := g.PacketHistory[origin]; !ok {
		// unknown origin, add it to wantlist
		g.WantList[origin] = uint32(1)

		g.PacketHistory[origin] = make(map[uint32]u.GossipPacket)
	}

	if _, ok := g.PacketHistory[origin][id]; ok {
		// packet already written
		g.HistoryMutex.Unlock()

		return false
	}
	l := len(g.PacketHistory[origin])

	// write packet to history
	g.PacketHistory[origin][id] = gp

	g.HistoryMutex.Unlock()

	if int(id) <= l {
		// filling missing slot
		g.WantListMutex.Lock()
		oldID, ok := g.WantList[origin]
		g.WantListMutex.Unlock()
		if !ok {
			g.Printer.Println("Fatal: history write error, uncomplete wantlist")
		}
		ok = true
		g.HistoryMutex.Lock()
		for ok {
			// to get first missing id
			_, ok = g.PacketHistory[origin][oldID]
			oldID++
		}
		g.HistoryMutex.Unlock()
		// update nextID
		oldID--
		g.WantListMutex.Lock()
		g.WantList[origin] = oldID
		g.WantListMutex.Unlock()

	} else if int(id) == l+1 {
		// next gossip
		// update wantlist
		g.WantListMutex.Lock()
		g.WantList[origin] = id + 1
		g.WantListMutex.Unlock()
	} else if int(id) > l+1 {
		// out of order (early)

		// do nothing
	} else {
		if g.ShouldPrint(logHW3, 1) {
			g.Printer.Println("Cannot write to gossip history, wrong id")
		}
	}

	// update new messages history (for GUI)
	if gp.Rumor != nil {
		if gp.Rumor.Text != "" {
			g.NewMessages.Mutex.Lock()
			g.NewMessages.Messages = append(g.NewMessages.Messages, *gp.Rumor)
			g.NewMessages.Mutex.Unlock()
		}
	}

	return true
}

// GetPeerID : returns the name of the peer
func (g *Gossiper) GetPeerID() string {
	return g.Name
}

// GetTLCRound GetTLCRound
func (g *Gossiper) GetTLCRound() int {
	return g.TLCRounds[g.Name]
}

// GetNewMessages : return the new unread messages
func (g *Gossiper) GetNewMessages(c int) []u.RumorMessage {
	g.NewMessages.Mutex.Lock()
	messages := g.NewMessages.Messages
	g.NewMessages.Mutex.Unlock()

	if c >= len(messages) {
		return nil
	}
	return messages[c:]
}

// CheckSearchFileComplete check if a search file is complete and update the
// corresponding field if it is
func (g *Gossiper) CheckSearchFileComplete(sf *u.SearchFile) {
	if sf.Complete {
		return
	}
	complete := true
	// iterate over all possible chunks
	for i := uint64(0); i < sf.NChunks; i++ {
		if _, ok := sf.Chunks[i]; !ok {
			// if one is missing, not complete
			complete = false
		}
	}
	if complete {
		sf.Complete = true
		g.SearchResults = append(g.SearchResults, *sf)
	}
}

// GetBlockchainMessageHistory for GUI
func (g *Gossiper) GetBlockchainMessageHistory(n int) []string {
	if n >= len(g.GUIlogHistory) {
		return nil
	}
	return g.GUIlogHistory[n:]
}

// GetCommittedHistory GetCommittedHistory
func (g *Gossiper) GetCommittedHistory(n int) []string {
	if n >= g.CommittedHistory.Size-1 {
		return nil
	}
	toRet := make([]string, g.CommittedHistory.Size-n-1)
	currNode := g.CommittedHistory.Tail
	for i := g.CommittedHistory.Size - n - 2; i >= 0; i-- {
		toRet[i] = currNode.Block.Transaction.Name
		currNode = currNode.Prev
	}
	return toRet
}
