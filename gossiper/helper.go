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
	if c >= u.SyncMapCount(g.Routes) {
		g.RouteMutex.Unlock()
		return nil
	}
	var dests []string
	// append all keys (string) of routes to dests
	f := func(k, v interface{}) bool {
		dests = append(dests, k.(string))
		return false
	}
	g.Routes.Range(f)
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

// HistoryMessageToByte : recover a rumor message from history, protobuf it
// and sends back an array of bytes
func (g *Gossiper) HistoryMessageToByte(ref u.MessageReference) []byte {
	rumor := g.RecoverHistoryRumor(ref)
	// protobuf the rumor to get a byte array
	packet := u.ProtobufGossip(&u.GossipPacket{Rumor: &rumor})
	return packet
}

// WriteRumorToHistory : write a given message to Gossiper message history
// return false if message already known, true otherwise
func (g *Gossiper) WriteRumorToHistory(rumor u.RumorMessage) bool {
	origin := rumor.Origin
	ID := rumor.ID
	text := rumor.Text

	// discard messages without origin or with message ID < 1
	if origin == "" {
		return false
	}

	if originHistory, ok := g.RumorHistory.Load(origin); !ok {
		var empty []u.HistoryMessage
		//g.HistoryMutex.Lock()
		g.RumorHistory.Store(origin, empty)
		//g.HistoryMutex.Unlock()

		//g.WantListMutex.Lock()
		g.WantList.Store(origin, uint32(1))
		//g.WantListMutex.Unlock()
	} else {
		for _, v := range originHistory.([]u.HistoryMessage) {
			if v.ID == ID {
				// message already written to history
				return false
			}
		}
	}

	// write message to history
	//g.HistoryMutex.Lock()
	oH, _ := g.RumorHistory.Load(origin)
	//g.HistoryMutex.Unlock()
	originHistory := oH.([]u.HistoryMessage)
	//g.WantListMutex.Lock()
	newID, _ := g.WantList.Load(origin)
	//g.WantListMutex.Unlock()
	if int(ID) < len(originHistory) { // packet received not in order
		// more recent messages have been received, but message missing

		// fill a missing slot
		originHistory[ID-1] = u.HistoryMessage{ID: ID, Text: text}

		found := false
		for i, v := range originHistory {
			// check for empty slots in message history to define the wantlist
			if v.ID == 0 {
				//g.WantListMutex.Lock()
				g.WantList.Store(origin, uint32(i+1))
				//g.WantListMutex.Unlock()
				found = true
				break
			}
		}
		// if not, the last slot has been filled, set wantlist value to the end
		// of the history (next value)
		if !found {
			//g.WantListMutex.Lock()
			g.WantList.Store(origin, uint32(len(originHistory)+1))
			//g.WantListMutex.Unlock()
		}
	} else if ID == newID {
		// the next packet that g doesn't have
		originHistory = append(originHistory,
			u.HistoryMessage{ID: ID, Text: text})
		//g.HistoryMutex.Lock()
		g.RumorHistory.Store(origin, originHistory)
		//g.HistoryMutex.Unlock()

		// update the wantlist to the next message
		//g.WantListMutex.Lock()
		curr, _ := g.WantList.Load(origin)
		g.WantList.Store(origin, uint32(curr.(uint32)+1))
		//g.WantListMutex.Unlock()
	} else { // not the wanted message, but a newer one
		// e.g waiting for message 2, and get message 4
		for i, _ := g.WantList.Load(origin); i.(uint32) < ID; i =
			i.(uint32) + 1 {
			// create empty slots for missing message
			originHistory = append(originHistory,
				u.HistoryMessage{ID: 0, Text: ""})
		}
		originHistory = append(originHistory,
			u.HistoryMessage{ID: ID, Text: text})
		//g.HistoryMutex.Lock()
		g.RumorHistory.Store(origin, originHistory)
		//g.HistoryMutex.Unlock()
		// don't update the wantlist, as wanted message still missing
	}

	if rumor.Text != "" {
		g.NewMessages.Mutex.Lock()
		defer g.NewMessages.Mutex.Unlock()
		g.NewMessages.Messages = append(g.NewMessages.Messages, rumor)
	}
	return true
}

// GetPeerID : returns the name of the peer
func (g *Gossiper) GetPeerID() string {
	return g.Name
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

// GetLastIDFromOrigin returns the last message ID received from origin
func (g *Gossiper) GetLastIDFromOrigin(origin string) uint32 {
	// load the history for the given origin
	//g.HistoryMutex.Lock()
	v, ok := g.RumorHistory.Load(origin)
	//g.HistoryMutex.Unlock()
	if !ok {
		// if author not found in history return 0
		return 0
	}
	// if found, we return the length of the rumor history associated with
	// the origin, which is equal to the last message id
	return uint32(len(v.([]u.HistoryMessage)))
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
