package utils

import (
	"crypto/sha256"
	"os"
	"time"
)

const (
	// SimpleModeStr : string containing the name for gossiper simple mode
	SimpleModeStr string = "simple"

	// RumorModeStr : string containing the name for gossiper rumor mode
	RumorModeStr string = "rumor"

	// TimeoutValue : timeout value in seconds
	TimeoutValue int = 10

	// AntiEntropyDefault : default value in seconds for anti entropy
	AntiEntropyDefault int = 10

	// RTimerDefault is the defaut value for rtimer
	RTimerDefault int = 0

	// LocalhostAddr : "127.0.0.1"
	LocalhostAddr string = "127.0.0.1"

	// DefaultUIPort : 8080
	DefaultUIPort string = "8080"

	// DefaultGUIPort : 8080
	DefaultGUIPort string = "8080"

	// PrivateMessageID is the ID for private messages
	PrivateMessageID uint32 = 0

	// DefaultHopLimit default hop limit for private messages
	DefaultHopLimit uint32 = 10

	// SharedFolderName path to the gossiper's shared folder
	SharedFolderName string = "_SharedFiles"

	// DownloadsFolderName path to the gossiper's downloads folder
	DownloadsFolderName string = "_Downloads"

	//DownloadsFolderPath string = SharedFolderPath

	// Filemode is file mode for the shared and download folders
	Filemode os.FileMode = 0666

	// ChunkSize size of a file chunk to send (8KB)
	ChunkSize int = 8192 // 8 KB

	// BufferSize : read buffer size in bytes
	BufferSize int = ChunkSize + 2048

	// ShaSize length of a sha hash in bytes
	ShaSize int = sha256.Size

	// FileTimout timeout value for file requests in seconds
	FileTimout int = 5

	// DefaultSearchBudget default search budget
	DefaultSearchBudget uint64 = 2

	// NoSearchBudget default search budget if not given as argument
	NoSearchBudget int = 0

	// MaxSearchBudget max search budget
	MaxSearchBudget uint64 = 32

	// DefaultGossiperName default gossiper name
	DefaultGossiperName string = "Gossiper"

	// DefaultDuplicateSearchTime default time in which the same search request
	// is considered as duplicate
	DefaultDuplicateSearchTime time.Duration = 500 * time.Millisecond

	// SearchPeriod search requests should be sent periodically with this
	// period
	SearchPeriod time.Duration = time.Second

	// MatchThreshold threshold of search matchs
	MatchThreshold int = 2

	// DefaultPeerNumber default N
	DefaultPeerNumber int = 3

	// StubbornTimeoutDefault in seconds
	StubbornTimeoutDefault int = 5
)
