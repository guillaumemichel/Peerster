package utils

import (
	"sync"
)

// Message : simple message type that client send to gossiper
type Message struct {
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
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
	Simple      *SimpleMessage
	Rumor       *RumorMessage
	Status      *StatusPacket
	Private     *PrivateMessage
	DataRequest *DataRequest
	DataReply   *DataReply
}

// PrivateMessage : private message structure
type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
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

// SyncNewMessages : sync slice of rumor messages
type SyncNewMessages struct {
	sync.Mutex
	Messages []RumorMessage
}

// ShaHash is the struct of hash Sha256
type ShaHash [ShaSize]byte

// FileChunk one chunk of data of a file
type FileChunk struct {
	File   *FileStruct
	Number int
	Hash   ShaHash
	Data   []byte
}

// FileStruct file structure containing file's metadata stored on the gossiper
type FileStruct struct {
	Name         string
	MetafileHash ShaHash
	Size         int64
	Metafile     []byte
	NChunks      int
	Chunks       map[ShaHash]*FileChunk
}

// DataRequest data request packets
type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

// DataReply data reply packets
type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

// FileRequestStatus status for a file request
type FileRequestStatus struct {
	sync.Mutex
	Name          string
	Destination   string
	MetafileHash  ShaHash
	MetafileOK    bool
	PendingChunks []ShaHash
	ChunkCount    int
	Data          [][]byte
	Ack           chan bool
}
