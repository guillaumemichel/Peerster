package gossiper

import (
	"encoding/hex"
	"fmt"

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
						data = fstatus.Data[i]
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

	f.CheckDownloadDir()
	var h u.ShaHash
	copy(h[:], drep.HashValue)

	for _, v := range g.FileStatus {
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

				return
			}
			v.MetafileOK = true
			v.PendingChunks = make([]u.ShaHash, 0)
			for i := 0; i < len(drep.Data); i += u.ShaSize {
				// add all hashes to pending chunks of v
				copy(h[:], drep.Data[i:i+u.ShaSize])
				v.PendingChunks = append(v.PendingChunks, h)
			}
			//g.Printer.Println(len(v.PendingChunks), "chunks in total")
			// request first chunk
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
					v.Data = append(v.Data, drep.Data)
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
		chunk := u.FileChunk{
			File:   &file,
			Number: i,
			Hash:   v,
			Data:   fstatus.Data[i],
		}
		// map the hash of the chunk to the chunk
		chunkMap[v] = &chunk
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
	// build the filestruct
	fstruct, err := f.ScanFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	// add it to the gossiper
	g.FileStructs = append(g.FileStructs, *fstruct)
	g.PrintHashOfIndexedFile(filename,
		hex.EncodeToString(fstruct.MetafileHash[:]))
}
