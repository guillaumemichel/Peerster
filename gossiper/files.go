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
	for _, c := range g.Chunks {
		if hash == c.Hash {
			//g.Printer.Println("\n", dreq.Origin)
			data = c.Data
			found = true
			break
		}
	}
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
			if hash == fstatus.MetafileHash {
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
						data = fstatus.Data[i].Data
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
	/*if !found {
		// hash not found
		g.Printer.Println("Warning: data requested and not found")
		return
	}*/
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
		if !v.MetafileOK && h == v.MetafileHash {
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
			v.PendingChunks = make([]u.ShaHash, 0)
			g.Chunks = append(g.Chunks, u.FileChunk{
				Hash: h,
				Data: drep.Data,
			})

			for i := 0; i < len(drep.Data); i += u.ShaSize {
				// add all hashes to pending chunks of v
				copy(h[:], drep.Data[i:i+u.ShaSize])
				v.PendingChunks = append(v.PendingChunks, h)
			}
			//g.Printer.Println(len(v.PendingChunks), "chunks in total")
			// request first chunk
			// write the chunk to the list of chunks

			g.RequestNextChunk(v)
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
					g.Chunks = append(g.Chunks, u.FileChunk{
						Number: v.ChunkCount,
						Hash:   h,
						Data:   drep.Data,
					})
					v.Data = append(v.Data, &g.Chunks[len(g.Chunks)-1])
					// increase chunk counter
					v.ChunkCount++
					if v.ChunkCount == len(v.PendingChunks) {
						// success rebuild the file
						g.ReconstructFile(v)
					} else {
						// request next chunk
						g.RequestNextChunk(v)
					}
					return
				}
			}
		}
	}
	// not wanted! register the file
	copy(h[:], drep.HashValue)
	metaf := make([]u.ShaHash, 0)
	for i := 0; i < len(drep.Data); i += u.ShaSize {
		// add all hashes to pending chunks of v
		copy(h[:], drep.Data[i:i+u.ShaSize])
		metaf = append(metaf, h)
	}
	c := make(chan bool)

	fstatus := u.FileRequestStatus{
		MetafileHash:  h,
		PendingChunks: metaf,
		MetafileOK:    true,
		ChunkCount:    -1, // -1 indicate that we received spontaneously a mfile
		Ack:           c,
	}
	g.FileStatus = append(g.FileStatus, &fstatus)
	g.Printer.Println("RECEIVED metafile", hex.EncodeToString(drep.HashValue),
		"from", drep.Origin)
	//g.Printer.Println("Warning: received data reply that I don't want!")
}

// ReconstructFile reconstruct a file after received all the chunks
func (g *Gossiper) ReconstructFile(fstatus *u.FileRequestStatus) {
	// translate pending chunks to metafile
	meta := make([]byte, 0)
	for _, v := range fstatus.PendingChunks {
		meta = append(meta, v[:]...)
	}

	// create he filestruct
	file := u.FileStruct{
		Name:         fstatus.Name,
		MetafileHash: fstatus.MetafileHash,
		Metafile:     meta,
		NChunks:      fstatus.ChunkCount,
	}

	// create the chunk map
	chunkMap := make(map[u.ShaHash]*u.FileChunk)
	for i, v := range fstatus.PendingChunks {
		// create a chunk for the chunk map
		/*
			chunk := u.FileChunk{
				File:   &file,
				Number: i,
				Hash:   v,
				Data:   fstatus.Data[i].Data,
			}
			// map the hash of the chunk to the chunk
			chunkMap[v] = &chunk
		*/
		chunkMap[v] = fstatus.Data[i]
	}

	file.Chunks = chunkMap

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
	file.Size = f.WriteFileToDownloads(&file)
	g.FileStructs = append(g.FileStructs, file)
}

// SendFileTo sends a file to dest
func (g *Gossiper) SendFileTo(dest, filename string) {
	//resolve file
	hash := make([]byte, 0)
	found := false
	// look in files stored on peer
	for _, fstruct := range g.FileStructs {
		if fstruct.Name == filename {
			hash = fstruct.MetafileHash[:]
			found = true
			break
		}
	}
	// look in files being downloaded by the peer
	if !found {
		for _, fstatus := range g.FileStatus {
			if fstatus.Name == filename {
				hash = fstatus.MetafileHash[:]
				found = true
				break
			}
		}
	}
	// if not found print error and abort
	if !found {
		g.Printer.Println("Error: file", filename, "not found!")
		return
	}

	// create data request
	dreq := u.DataRequest{
		Origin:      dest,
		Destination: g.Name,
		HopLimit:    u.DefaultHopLimit,
		HashValue:   hash,
	}
	// handle the data request, this will generate the corresponding data reply
	g.HandleDataReq(dreq)
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
	// add it to the gossiper
	g.FileStructs = append(g.FileStructs, *fstruct)
	g.PrintHashOfIndexedFile(filename,
		hex.EncodeToString(fstruct.MetafileHash[:]))
}

// ScanFile scans a file and split it into chunks
func (g *Gossiper) ScanFile(f *os.File) (*u.FileStruct, error) {
	// get basic file infos
	fstat, _ := f.Stat()

	// create the file structure
	filestruct := u.FileStruct{
		Name: fstat.Name(),
		Size: fstat.Size(),
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
		g.Chunks = append(g.Chunks, chunk)

		// associate the hash with the chunk
		chunks[hash] = &g.Chunks[len(g.Chunks)-1]
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
