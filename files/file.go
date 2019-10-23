package files

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	u "github.com/guillaumemichel/Peerster/utils"
)

// LoadSharedFiles Load all files from _SHAREDFILES and return them, in order
// to store them in the gossiper
func LoadSharedFiles() []u.FileStruct {
	// TODO
	return make([]u.FileStruct, 0)
}

// LoadFile get a file from the filename
func LoadFile(filename string) *os.File {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	return f
}

// ScanFile scans a file and split it into chunks
func ScanFile(f os.File) (*u.FileStruct, map[u.ShaHash]*u.FileChunk, error) {
	// get basic file infos
	fstat, _ := f.Stat()

	// create the file structure
	filestruct := u.FileStruct{
		Name: fstat.Name(),
		Size: fstat.Size(),
	}

	// create the chunk map
	var metafile []byte
	var chunks map[u.ShaHash]*u.FileChunk
	tmp := make([]byte, u.ChunkSize)
	chunkCount := 0

	for {
		n, err := f.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
				return nil, nil, err
			}
			// end of file, last chunk
			break
		}
		chunkCount++

		// hash the read chunk
		hash := sha256.Sum256(tmp[:n])
		// add the hash to metafile
		metafile = append(metafile, hash[:]...)

		var data []byte
		copy(data, tmp)
		// create the file chuck
		chunk := &u.FileChunk{
			File:   &filestruct,
			Number: chunkCount,
			Hash:   hash,
			Data:   data,
		}
		// associate the hash with the chunk
		chunks[hash] = chunk
	}

	if u.ChunkSize < len(metafile) {
		fmt.Println("Error: metafile too large")
	}

	// update the filestruct with metafile information and chunk count
	filestruct.Metafile = metafile
	filestruct.MetafileHash = sha256.Sum256(metafile)
	filestruct.NChunks = chunkCount

	return &filestruct, chunks, nil
}
