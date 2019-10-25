package gossiper

import (
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
	f.CheckDownloadDir()
	var h u.ShaHash
	copy(h[:], drep.HashValue)

	for _, v := range g.FileStatus {
		// a metafile we were waiting for is here !
		if !v.MetafileOK && h == v.MetafileHash {
			// ack the metafile
			v.MetafileOK = true
			v.PendingChunks = make([]u.ShaHash, 0)
			for i := 0; i < len(drep.Data); i += u.ShaSize {
				// add all hashes to pending chunks of v
				copy(h[:], drep.Data[i:i+u.ShaSize])
				v.PendingChunks = append(v.PendingChunks, h)
			}
			// request first chunk
			g.RequestNextChunk(v)
			return
		}
		if v.MetafileOK { // if metafile ok, we look for a chunk
			// we iterate over all pending chunks
			for _, w := range v.PendingChunks {
				// if the hash matches
				if h == w {
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
	g.Printer.Println("Warning: received data reply that I don't want!")
}

// ReconstructFile reconstruct a file after received all the chunks
func (g *Gossiper) ReconstructFile(fstatus *u.FileRequestStatus) {
	// translate pending chunks to metafile
	meta := make([]byte, 0)
	for _, v := range fstatus.PendingChunks {
		meta = append(meta, v[:]...)
	}

	chunkMap := make(map[u.ShaHash]*u.FileChunk)
	// TODO fill map
	// create he filestruct
	file := u.FileStruct{
		Name:         fstatus.Name,
		MetafileHash: fstatus.MetafileHash,
		Metafile:     meta,
		NChunks:      fstatus.ChunkCount,
		Chunks:       chunkMap,
	}

	g.FileStructs = append(g.FileStructs, file)
	file.Size = f.WriteFileToDownloads(&file)
}

// SendFileTo sends a file to dest
func (g *Gossiper) SendFileTo(dest, file string) {
	//resolve file
	hash := make([]byte, 0)
	dreq := u.DataRequest{
		Origin:      dest,
		Destination: g.Name,
		HopLimit:    u.DefaultHopLimit,
		HashValue:   hash,
	}
	g.HandleDataReq(dreq)
}
