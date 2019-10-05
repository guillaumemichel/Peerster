package types

// SimpleMessage : a simple message structure containing the original sender's
// name, the relay peer address, and the content of the message
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

// GossipPacket : default gossip packet used by the Peerster
type GossipPacket struct {
	Simple *SimpleMessage
}
