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
	Budget      *uint64
	Keywords    *[]string
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
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
	TLCMessage    *TLCMessage
	Ack           *TLCAck
}

// PrivateMessage : private message structure
type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

// AckValues : values pointed by an AckIdentifier
type AckValues struct {
	Channel        *chan bool
	InitialMessage GossipPacket
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
	Done         bool
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
	File          *FileStruct
	Destination   [][]string
	MetafileOK    bool
	PendingChunks []ShaHash
	ChunkCount    int
	Ack           chan bool
}

// FileStatusList list of file request statuses
type FileStatusList struct {
	sync.Mutex
	List []FileRequestStatus
}

// SearchRequest search request packet
type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

// SearchReply search reply packet
type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

// SearchResult search result included in search reply
type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

// SearchStatus search status kept at the gossiper
type SearchStatus struct {
	Origin   string
	Keywords map[string]bool
}

// SearchFile search file structure
type SearchFile struct {
	Name         string
	MetafileHash ShaHash
	Chunks       map[uint64]map[string]bool
	NChunks      uint64
	Complete     bool
}

// TxPublish transaction publish structure
type TxPublish struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

// BlockPublish block publish structure
type BlockPublish struct {
	PrevHash    [32]byte
	Transaction TxPublish
}

// TLCMessage structure
type TLCMessage struct {
	Origin      string
	ID          uint32
	Confirmed   int
	TxBlock     BlockPublish
	VectorClock *StatusPacket
	Fitness     float32
}

// TLCAck TLC message acknowledgement
type TLCAck PrivateMessage
