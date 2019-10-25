package files

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	u "github.com/guillaumemichel/Peerster/utils"
)

// LoadSharedFiles Load all files from _SHAREDFILES and return them, in order
// to store them in the gossiper
func LoadSharedFiles() []u.FileStruct {
	// read the shared files directory
	files, err := ioutil.ReadDir(u.SharedFolderPath)
	if err != nil {
		log.Fatal(err)
	}

	// we assume that there is no subdirectory
	var fileStructs []u.FileStruct
	// iterate over the files in this directory
	for _, f := range files {
		// create the filestruct from the filename
		fs, err := ScanFile(LoadFile(f.Name()))
		if err != nil {
			log.Fatal(err)
			return nil
		}
		fileStructs = append(fileStructs, *fs)
	}

	return fileStructs
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
func ScanFile(f *os.File) (*u.FileStruct, error) {
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
				return nil, err
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
	filestruct.Chunks = chunks

	return &filestruct, nil
}

// WriteFileToDownloads write a file to _Downloads folder and returns the size
// of the created file
func WriteFileToDownloads(fstruct *u.FileStruct) int64 {
	// loading data
	data := make([]byte, 0)
	var h u.ShaHash
	for i := 0; i < len(fstruct.Metafile); i += u.ShaSize {
		// translate hash []byte to shahash (in h)
		copy(h[:], fstruct.Metafile[i:i+u.ShaSize])
		// load the data from the chunk in the same order
		data = append(data, fstruct.Chunks[h].Data...)
	}

	// writing the file
	err := ioutil.WriteFile(u.DownloadsFolderPath+"/"+fstruct.Name,
		data, u.Filemode)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return int64(len(data))
}

func checkDir(dir string) {
	// make sure that the shared files folder exists, and create it if it does
	// not exist yet
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(u.SharedFolderPath, u.Filemode)
		} else {
			fmt.Println("Error: cannot open", dir)
		}
	}
}

// CheckDownloadDir check if download directory exists, and create it if it does
// not exist
func CheckDownloadDir() {
	checkDir(u.DownloadsFolderPath)
}

// CheckSharedDir check if shared files directory exists, and create it if it
// does not exist
func CheckSharedDir() {
	checkDir(u.SharedFolderPath)
}
