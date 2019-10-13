package utils

// Message : simple message type that client send to gossiper
type Message struct {
	Text string
}

// SimpleMessage : a simple message structure containing the original sender's
// name, the relay peer address, and the content of the message
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

// RumorMessage : Rumor message containing the original sender's name, a
// sequence number and the content. The message should be uniquely identified
// by the combination of Origin and ID
type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

// PeerStatus : a status is useful to ask a peer some packets
type PeerStatus struct {
	Identifier string
	NextID     uint32
}

// StatusPacket : a StatusPacket is a packet containing a list of PeerStatus
// that are wanted by the sender
type StatusPacket struct {
	Want []PeerStatus
}

// GossipPacket : default gossip packet used by the Peerster
type GossipPacket struct {
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
}

// HistoryMessage : message stored in Gossiper history in an arry identifying
// the origin of those messages
type HistoryMessage struct {
	ID   uint32
	Text string
}

// MessageReference : reference to a message stored in history
type MessageReference struct {
	Origin string
	ID     uint32
}

// AckIdentifier : Ack/status identifier for a sent message
type AckIdentifier struct {
	Peer   string
	Origin string
	ID     uint32
}

// AckValues : values pointed by an AckIdentifier
type AckValues struct {
	Channel        chan bool
	InitialMessage MessageReference
}
