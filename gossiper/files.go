package gossiper

import (
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

}
