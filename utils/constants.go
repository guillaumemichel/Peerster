package utils

import (
	"crypto/sha256"
	"os"
)

// SimpleModeStr : string containing the name for gossiper simple mode
const SimpleModeStr string = "simple"

// RumorModeStr : string containing the name for gossiper rumor mode
const RumorModeStr string = "rumor"

// TimeoutValue : timeout value in seconds
const TimeoutValue int = 10

// AntiEntropyDefault : default value in seconds for anti entropy
const AntiEntropyDefault int = 10

// RTimerDefault is the defaut value for rtimer
const RTimerDefault int = 0

// BufferSize : read buffer size in bytes
const BufferSize int = 2048

// LocalhostAddr : "127.0.0.1"
const LocalhostAddr string = "127.0.0.1"

// DefaultUIPort : 8080
const DefaultUIPort string = "8080"

// DefaultGUIPort : 8080
const DefaultGUIPort string = "8080"

// PrivateMessageID is the ID for private messages
const PrivateMessageID uint32 = 0

// DefaultHopLimit default hop limit for private messages
const DefaultHopLimit uint32 = 10

// SharedFolderPath path to the gossiper's shared folder
const SharedFolderPath string = "./_SharedFiles"

// DownloadsFolderPath path to the gossiper's downloads folder
const DownloadsFolderPath string = "./_Downloads"

// Filemode is file mode for the shared and download folders
const Filemode os.FileMode = 0777

// ChunkSize size of a file chunk to send (8KB)
const ChunkSize int = 8192 // 8 KB

// ShaSize length of a sha hash in bytes
const ShaSize int = sha256.Size

// FileTimout timeout value for file requests in seconds
const FileTimout int = 5
