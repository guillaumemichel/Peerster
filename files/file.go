package files

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	u "github.com/guillaumemichel/Peerster/utils"
)

/*
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
*/

// LoadFile get a file from the filename
func LoadFile(filename string) *os.File {
	base, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	path := filepath.Join(base, u.SharedFolderName, filename)
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return f
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
		// fmt.Println(h)
		//fmt.Println(string(fstruct.Chunks[h].Data))
		if h != sha256.Sum256(fstruct.Chunks[h].Data) {
			fmt.Println("mismatch between data and hash")
		}
		data = append(data, fstruct.Chunks[h].Data...)
	}

	base, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return 0
	}

	// writing the file
	path := filepath.Join(base, u.DownloadsFolderName, fstruct.Name)
	err = ioutil.WriteFile(path, data, u.Filemode)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

func checkDir(dir string) {
	// make sure that the shared files folder exists, and create it if it does
	// not exist yet
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, u.Filemode)
		} else {
			fmt.Println("Error: cannot open", dir)
		}
	}
}

// CheckDownloadDir check if download directory exists, and create it if it does
// not exist
func CheckDownloadDir() {
	base, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	path := filepath.Join(base, u.DownloadsFolderName)
	checkDir(path)
}

// CheckSharedDir check if shared files directory exists, and create it if it
// does not exist
func CheckSharedDir() {
	base, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	path := filepath.Join(base, u.SharedFolderName)
	checkDir(path)
}
