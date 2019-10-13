package gossiper

import (
	"log"

	u "github.com/guillaumemichel/Peerster/utils"
)

// CreateSimpleMessage : creates a gossip simple message from a string
func (g *Gossiper) CreateSimpleMessage(name, content string) u.SimpleMessage {
	return u.SimpleMessage{
		OriginalName:  name,
		RelayPeerAddr: g.GossipAddr.String(),
		Contents:      content,
	}
}

// CreateRumorMessage :creates a gossip rumor message from a string, and
// increases the rumorCount from the Gossiper
func (g *Gossiper) CreateRumorMessage(content string) u.RumorMessage {
	g.RumorCount++
	return u.RumorMessage{
		Origin: g.Name,
		ID:     g.RumorCount,
		Text:   content,
	}
}

// ReplaceRelayPeerSimple : replaces the relay peer of a simple message with its
// own address
func (g *Gossiper) ReplaceRelayPeerSimple(
	msg *u.SimpleMessage) *u.SimpleMessage {

	msg.RelayPeerAddr = g.GossipAddr.String()
	if msg.RelayPeerAddr == "<nil>" {
		log.Fatal("cannot replace relay peer address")
	}
	return msg
}

// BuildStatusPacket : build a status packet for g
func (g *Gossiper) BuildStatusPacket() u.StatusPacket {

	var want []u.PeerStatus

	f := func(k, v interface{}) bool {
		want = append(want, u.PeerStatus{Identifier: k.(string),
			NextID: v.(uint32)})
		return true
	}

	g.WantList.Range(f)
	sp := u.StatusPacket{Want: want}
	return sp
}
