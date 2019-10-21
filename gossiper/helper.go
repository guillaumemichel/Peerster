package gossiper

import (
	"log"
	"net"
	"strings"

	u "github.com/guillaumemichel/Peerster/utils"
)

// GetRandPeer : get a random peer known to g
func (g *Gossiper) GetRandPeer() *net.UDPAddr {
	if len(g.Peers) == 0 {
		//fmt.Println("Error: cannot send message to random peer,",
		//	"no known peer :(")
		return nil
	}
	r := u.GetRealRand(len(g.Peers)) // GetRand(n) also possible
	target := (g.Peers)[r]
	return &target
}

// GetPeers : return a list of string all peer addresses
func (g *Gossiper) GetPeers() []string {
	var list []string
	for _, v := range g.Peers {
		list = append(list, v.String())
	}
	return list
}

// AddPeer : adds the given peer to peers list if not already in it
func (g *Gossiper) AddPeer(addr *net.UDPAddr) {
	addrStr := addr.String()
	if !strings.Contains(g.PeersToString(), addrStr) &&
		addrStr != g.GossipAddr.String() && addrStr != g.GossipAddr.String() {
		g.Peers = append(g.Peers, *addr)
	}
}

// ReceiveOK : return false if message too long to be handled, true otherwise
func (g *Gossiper) ReceiveOK(ok bool, rcvBytes []byte) bool {
	if !ok && len(rcvBytes) >= g.BufSize {
		log.Printf(`Warning: incoming message possibly larger than %d bytes 
			couldn't be read!\n`, g.BufSize)
		return false
	}
	return true
}

// RecoverHistoryRumor : recover a message from rumor history
func (g *Gossiper) RecoverHistoryRumor(ref u.MessageReference) u.RumorMessage {

	oH, ok := g.RumorHistory.Load(ref.Origin)
	// if ID ref is larger than array size or message not in array
	if !ok || len(oH.([]u.HistoryMessage)) < int(ref.ID) ||
		oH.([]u.HistoryMessage)[ref.ID-1].ID != ref.ID {

		log.Println("Error: message queried and not found in history!")
		log.Println(ref.Origin, " : ", ref.ID, " : ",
			len(oH.([]u.HistoryMessage)))
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
	if origin == "" || rumor.ID < 1 {
		return false
	}

	if originHistory, ok := g.RumorHistory.Load(origin); !ok {
		var empty []u.HistoryMessage
		g.RumorHistory.Store(origin, empty)

		g.WantList.Store(origin, uint32(1))
	} else {
		for _, v := range originHistory.([]u.HistoryMessage) {
			if v.ID == ID {
				// message already written to history
				return false
			}
		}
	}

	// write message to history
	oH, _ := g.RumorHistory.Load(origin)
	originHistory := oH.([]u.HistoryMessage)
	if int(ID) < len(originHistory) { // packet received not in order
		// more recent messages have been received, but message missing

		// fill a missing slot
		originHistory[ID-1] = u.HistoryMessage{ID: ID, Text: text}

		found := false
		for i, v := range originHistory {
			// check for empty slots in message history to define the wantlist
			if v.ID == 0 {
				g.WantList.Store(origin, uint32(i+1))
				found = true
				break
			}
		}
		// if not, the last slot has been filled, set wantlist value to the end
		// of the history (next value)
		if !found {
			g.WantList.Store(origin, uint32(len(originHistory)+1))
		}
	} else if newID, _ := g.WantList.Load(origin); ID == newID {
		// the next packet that g doesn't have
		originHistory = append(originHistory,
			u.HistoryMessage{ID: ID, Text: text})
		g.RumorHistory.Store(origin, originHistory)

		// NOOOOOOOO MUTEX CAN BE SHITTY AROUND HERE

		// update the wantlist to the next message
		curr, _ := g.WantList.Load(origin)
		g.WantList.Store(origin, uint32(curr.(uint32)+1))
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
		g.RumorHistory.Store(origin, originHistory)
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
func (g *Gossiper) GetNewMessages() []u.RumorMessage {
	g.NewMessages.Mutex.Lock()
	defer g.NewMessages.Mutex.Unlock()
	return g.NewMessages.Messages
	/*
			var out []u.RumorMessage
		// get the new messages to out
		for _, v := range g.NewMessages.Messages {
			out = append(out, v)
		}
		// can be removed for complete history
		//g.NewMessages.Messages = make([]u.RumorMessage, 0)

		return out*/
}

// GetLastIDFromOrigin returns the last message ID received from origin
func (g *Gossiper) GetLastIDFromOrigin(origin string) uint32 {
	// load the history for the given origin
	v, ok := g.WantList.Load(origin)
	if !ok {
		// if author not found in history return 0
		return 0
	}
	// if found return the waited one minus one (the last received)
	return v.(uint32) - 1
}
