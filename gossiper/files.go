package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	f "github.com/guillaumemichel/Peerster/files"
	u "github.com/guillaumemichel/Peerster/utils"
)

// LookForMetahash look if gossiper has the file with the same metahash
func (g *Gossiper) LookForMetahash(target u.ShaHash) *u.FileStruct {
	// iterate over the filestructs of g
	for _, v := range g.FileStructs {
		// if same metahash, return the file
		if v.MetafileHash == target {
			return &v
		}
	}
	return nil
}

// HandleDataReq handles data requests that are received
func (g *Gossiper) HandleDataReq(dreq u.DataRequest) {
	var data []byte
	var hash u.ShaHash
	found := false
	copy(hash[:], dreq.HashValue)
	// iterating over known structs
	g.ChunkLock.Lock()
	for _, c := range g.Chunks {
		if hash == c.Hash {
			//g.Printer.Println("\n", dreq.Origin)
			data = c.Data
			found = true
			break
		}
	}
	g.ChunkLock.Unlock()

	if !found {
		for _, fstruct := range g.FileStructs {
			// if the wanted hash is metafilehash, data is the metafile
			if hash == fstruct.MetafileHash {
				data = fstruct.Metafile
				found = true
				break
				// else if the desired data is a block, data <- this block
			} else if v, ok := fstruct.Chunks[hash]; ok {
				data = v.Data
				found = true
				break
			}
		}
	}

	// if not found look in the files being downloaded
	if !found {
		for _, fstatus := range g.FileStatus {
			// if the wanted hash is metafilehash, data is the pending chunk
			if hash == fstatus.File.MetafileHash {
				// convert pending chunks to metafile
				for _, v := range fstatus.PendingChunks {
					data = append(data, v[:]...)
				}
				found = true
				break
				// else if the desired data is a block, data <- this block
			}
			for i, v := range fstatus.PendingChunks {
				// found block
				if hash == v {
					if i <= fstatus.ChunkCount {
						// if block is already downloaded
						//data = fstatus.Data[i].Data
						data = fstatus.File.Chunks[hash].Data
						found = true
						break
					} else {
						g.Printer.Println("Warning: data requested",
							"not downloaded yet")
						return
					}
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		// hash not found
		g.Printer.Println("Warning: data requested and not found")
		return
	}
	// create the data reply
	drep := u.DataReply{
		Origin:      g.Name,
		Destination: dreq.Origin,
		HopLimit:    u.DefaultHopLimit,
		HashValue:   dreq.HashValue,
		Data:        data,
	}
	// route it to the next hop
	g.RouteDataReply(drep)
}

// HandleDataReply handles data replies that are received
func (g *Gossiper) HandleDataReply(drep u.DataReply) {
	//g.Printer.Println("Got reply from", drep.Origin)

	if drep.Destination != g.Name {
		g.RouteDataReply(drep)
	}

	f.CheckDownloadDir()
	var h u.ShaHash
	copy(h[:], drep.HashValue)

	statuses := g.FileStatus
	for _, v := range statuses {
		// a metafile we were waiting for is here !
		if !v.MetafileOK && h == v.File.MetafileHash {
			// ack the metafile
			v.Ack <- true
			// check if peer had the file
			if len(drep.Data) == 0 {
				g.Printer.Println(drep.Origin, "doesn't have the file",
					hex.EncodeToString(drep.HashValue))
				// delete the filestatus from g
				fs := g.FileStatus
				for i, w := range fs {
					if w == v {
						// erase the element
						fs[i] = fs[len(fs)-1]
						fs[len(fs)-1] = nil
						fs = fs[:len(fs)-1]
					}
				}
				g.FileStatus = fs
				return
			}
			v.MetafileOK = true
			g.ChunkLock.Lock()
			g.Chunks = append(g.Chunks, u.FileChunk{
				File: v.File,
				Hash: h,
				Data: drep.Data,
			})
			g.ChunkLock.Unlock()

			v.PendingChunks = make([]u.ShaHash, 0)
			for i := 0; i < len(drep.Data); i += u.ShaSize {
				// add all hashes to pending chunks of v
				copy(h[:], drep.Data[i:i+u.ShaSize])
				v.PendingChunks = append(v.PendingChunks, h)
			}

			v.File.Metafile = drep.Data
			v.File.NChunks = len(v.PendingChunks)
			// file downloaded from an unique source
			if len(v.Destination) == 1 {
				// fill the destination array with this source for each chunk
				dests := make([][]string, len(v.PendingChunks))
				for i := 0; i < len(v.PendingChunks); i++ {
					dests[i] = []string{v.Destination[0][0]}
				}
				v.Destination = dests
			}

			// request first chunk
			// write the chunk to the list of chunks

			g.SendFileRequest(v, nil, v.Destination[v.ChunkCount])
			return
		}
		if v.MetafileOK { // if metafile ok, we look for a chunk
			// we iterate over all pending chunks
			for _, w := range v.PendingChunks {
				// if the hash matches
				if h == w {
					// ack the chunk
					v.Ack <- true
					// check if peer had the file
					if len(drep.Data) == 0 {
						g.Printer.Println(drep.Origin, "doesn't have the file",
							hex.EncodeToString(drep.HashValue))
						return
					}
					// append data to the one we already have
					g.ChunkLock.Lock()
					g.Chunks = append(g.Chunks, u.FileChunk{
						File:   v.File,
						Number: v.ChunkCount,
						Hash:   h,
						Data:   drep.Data,
					})
					v.File.Chunks[h] = &g.Chunks[len(g.Chunks)-1]
					g.ChunkLock.Unlock()
					//v.Data = append(v.Data, &g.Chunks[len(g.Chunks)-1])
					// increase chunk counter
					v.ChunkCount++
					if v.ChunkCount == len(v.PendingChunks) {
						// success rebuild the file
						g.ReconstructFile(v)
					} else {
						// request next chunk
						g.SendFileRequest(v, nil, v.Destination[v.ChunkCount])
					}
					return
				}
			}
		}
	}
}

// ReconstructFile reconstruct a file after received all the chunks
func (g *Gossiper) ReconstructFile(fstatus *u.FileRequestStatus) {
	// translate pending chunks to metafile

	fstatus.File.Done = true
	file := fstatus.File

	// delete the filestatus from g
	found := false
	fs := g.FileStatus
	for i, v := range fs {
		if v == fstatus {
			// erase the element
			fs[i] = fs[len(fs)-1]
			fs[len(fs)-1] = nil
			fs = fs[:len(fs)-1]
			found = true
		}
	}
	// could not remove the filestatus from g
	if !found {
		g.Printer.Println("Error: could not delete filestatus after successful",
			"download")
	}
	g.FileStatus = fs

	g.PrintReconstructFile(file.Name)

	// append the filestruct to known files
	file.Size = f.WriteFileToDownloads(file)
	//g.FileStructs = append(g.FileStructs, file)
}

// IndexFile indexes a file from the given filename and adds it to the gossiper
func (g *Gossiper) IndexFile(filename string) {
	// load the file from os
	file := f.LoadFile(filename)
	if file == nil {
		return
	}
	// build the filestruct
	fstruct, err := g.ScanFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range g.FileStructs {
		if filename == fstruct.Name && v.MetafileHash == fstruct.MetafileHash {
			g.Printer.Println("Warning: file already indexed. Aborting.")
			return
		}
	}
	// manage TLC
	if g.Hw3ex2 || g.Hw3ex3 || g.Hw3ex4 {
		if !g.ManageTLC(filename, fstruct.Size, fstruct.MetafileHash) {
			return
		}
	}

	// once it is confirmed, continue
	// may take some time to get acks

	// add it to the gossiper
	g.FileStructs = append(g.FileStructs, *fstruct)
	g.PrintHashOfIndexedFile(filename,
		hex.EncodeToString(fstruct.MetafileHash[:]), fstruct.NChunks)
}

// ScanFile scans a file and split it into chunks
func (g *Gossiper) ScanFile(f *os.File) (*u.FileStruct, error) {
	// get basic file infos
	fstat, _ := f.Stat()

	// create the file structure
	filestruct := u.FileStruct{
		Name: fstat.Name(),
		Size: fstat.Size(),
		Done: true,
	}

	// create the chunk map
	var metafile []byte
	chunks := make(map[u.ShaHash]*u.FileChunk)
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
		dat := make([]byte, n)
		copy(dat, tmp[:n])

		// create the file chuck
		chunk := u.FileChunk{
			File:   &filestruct,
			Number: chunkCount,
			Hash:   hash,
			Data:   dat,
		}
		g.ChunkLock.Lock()
		g.Chunks = append(g.Chunks, chunk)

		// associate the hash with the chunk
		chunks[hash] = &g.Chunks[len(g.Chunks)-1]
		g.ChunkLock.Unlock()
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

// HandleDownload download a file without specifying a destination
func (g *Gossiper) HandleDownload(filename string, request []byte) {
	var h u.ShaHash
	copy(h[:], request)
	for _, sr := range g.SearchResults {
		if g.ShouldPrint(logHW3, 3) {
			g.Printer.Println("Got hash:",
				hex.EncodeToString(sr.MetafileHash[:]), "request is:",
				hex.EncodeToString(h[:]))
		}
		if sr.MetafileHash == h {
			g.DownloadFile(filename, sr)
			return
		}
	}
	g.Printer.Println("Error: no search result match the given request")
}

// DownloadFile download a file from its search file results
func (g *Gossiper) DownloadFile(filename string, f u.SearchFile) {

	var fstruct *u.FileStruct
	for _, fs := range g.FileStructs {
		if fs.MetafileHash == f.MetafileHash {
			if fs.Done {
				g.Printer.Println("Warning: file already downloaded")
				return
			}
			fstruct = &fs
			fstruct.Name = f.Name
		}
	}

	chunks := make(map[u.ShaHash]*u.FileChunk)
	if fstruct == nil {
		fstruct = &u.FileStruct{
			Name:         f.Name,
			MetafileHash: f.MetafileHash,
			Done:         false,
			Chunks:       chunks,
		}
	}
	allDests := make(map[string]bool)

	// create destination array
	dests := make([][]string, f.NChunks)
	g.RouteMutex.Lock()
	for i := uint64(0); i < f.NChunks; i++ {
		// create the array for each chunk
		dests[i] = make([]string, 0)
		for k := range f.Chunks[i] {
			// if we have a route to dest, add it to dests
			if _, ok := g.Routes[k]; ok {
				dests[i] = append(dests[i], k)
				allDests[k] = true
			}
		}
	}
	g.RouteMutex.Unlock()

	c := make(chan bool)
	// create the new file status
	fstatus := u.FileRequestStatus{
		File:        fstruct,
		Destination: dests,
		MetafileOK:  false,
		ChunkCount:  0,
		Ack:         c,
	}
	g.FileStatus = append(g.FileStatus, &fstatus)

	mfdests := make([]string, len(allDests))
	count := 0
	for d := range allDests {
		mfdests[count] = d
		count++
	}

	g.SendFileRequest(&fstatus, &f.MetafileHash, mfdests)

}
