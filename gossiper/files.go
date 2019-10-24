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

}

// HandleDataReply handles data replies that are received
func (g *Gossiper) HandleDataReply(drep u.DataReply) {

}
