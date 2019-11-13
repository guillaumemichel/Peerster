package gossiper

import (
	"sort"
	"strings"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ManageSearch manage a search
func (g *Gossiper) ManageSearch(initialBudget uint64, keywords []string) {
	//TODO channels
	g.SendNewSearchReq(initialBudget, keywords)
}

// SendNewSearchReq creates and sends a new file search request
func (g *Gossiper) SendNewSearchReq(budget uint64, keywords []string) {
	// create the request
	req := u.SearchRequest{
		Origin:   g.Name,
		Budget:   budget,
		Keywords: keywords,
	}
	// handles it
	g.HandleSearchReq(req)
}

// HandleSearchReq handles a search request
func (g *Gossiper) HandleSearchReq(req u.SearchRequest) {
	// register request
	g.SearchMutex.Lock()
	// check if search request is duplicate
	if g.NewSearch(req) {
		// if not register it
		g.AddSearch(req)
		// and deletes it after a timeout
		go g.DeleteSearchAfterTimeout(req)
	}
	g.SearchMutex.Unlock()

	// search for file locally
	g.SearchForFile(req.Origin, req.Keywords)
	// substract 1 to the budget
	req.Budget--

	if req.Budget > 0 {
		quotient := int(req.Budget) / len(g.Peers)
		rest := int(req.Budget) % len(g.Peers)
		randPeers := u.ChooseCRandomAmongN(rest, len(g.Peers))
		if randPeers == nil {
			return
		}

		for i, p := range g.Peers {
			// set the newbudget for each request to budget/n_peers
			myBudget := quotient
			// if p was randomly chosen add 1 to its budget
			if u.IntInSlice(i, randPeers) {
				myBudget++
			}
			// if budget <= 0 don't send request
			if myBudget > 0 {
				// create the new search request
				newReq := u.SearchRequest{
					Origin:   req.Origin,
					Budget:   uint64(myBudget),
					Keywords: req.Keywords,
				}
				gp := u.GossipPacket{SearchRequest: &newReq}
				// route the request to next peer
				g.SendPacketToNeighbor(p, gp)
			}
		}
	}
}

// SearchForFile search for a file locally
func (g *Gossiper) SearchForFile(dest string, keywords []string) {
	// create a map from filename to possessed chunk numbers
	chunkMap := make(map[*u.FileStruct]map[int]bool)
	// iterate over all known chunks
	for _, c := range g.Chunks {
		// iterate over all keywords
		for _, w := range keywords {
			// if the name of the chunk correspond to a keyword
			if strings.Contains(c.File.Name, w) {
				// if filename's map doesn't exist yet
				if _, ok := chunkMap[c.File]; !ok {
					// initialize the map corresponding to the filename
					chunkMap[c.File] = make(map[int]bool)
				}
				// we have that chunk
				chunkMap[c.File][c.Number] = true
			}
		}
	}

	results := make([]*u.SearchResult, 0)
	for k, v := range chunkMap {
		// create the list of distinct chunk ids
		cmap := make([]uint64, 0)
		for i := range v {
			cmap = append(cmap, uint64(i))
		}
		// sort the slice in ascending order
		sort.Slice(cmap, func(i, j int) bool {
			return cmap[i] < cmap[j]
		})
		// create the search result
		res := u.SearchResult{
			FileName:     k.Name,
			MetafileHash: k.MetafileHash[:],
			ChunkMap:     cmap,
			ChunkCount:   uint64(k.NChunks),
		}
		// append it to the list
		results = append(results, &res)
	}
	if len(results) > 0 {
		g.SendSearchReply(dest, results)
	}
}

// SendSearchReply creates and sends search reply
func (g *Gossiper) SendSearchReply(dest string, results []*u.SearchResult) {
	rep := u.SearchReply{
		Origin:      g.Name,
		Destination: dest,
		HopLimit:    u.DefaultHopLimit,
		Results:     results,
	}
	g.HandleSearchReply(rep)
}

// HandleSearchReply handles search replies
func (g *Gossiper) HandleSearchReply(rep u.SearchReply) {
	//TODO if the reply is for me

	// decrease hop limit
	rep.HopLimit--
	gp := u.GossipPacket{SearchReply: &rep}
	// route the packet
	g.RoutePacket(rep.Destination, gp)
}

// NewSearch return true if the search request is new, false otherwise
func (g *Gossiper) NewSearch(req u.SearchRequest) bool {
	// g.SearchMutex should be locked
	for _, s := range g.SearchStatuses {
		// check if the search is in g status list
		if u.SameSearch(s, req) {
			return false
		}
	}

	return true
}

// AddSearch add a search status to g
func (g *Gossiper) AddSearch(req u.SearchRequest) {
	// g.SearchMutex should be locked

	// map a string to true if it is present in Keywords (not mapped otherwise)
	// act as a set
	kw := make(map[string]bool)
	for _, w := range req.Keywords {
		kw[w] = true
	}

	// create the search status
	search := u.SearchStatus{
		Origin:   req.Origin,
		Keywords: kw,
	}
	// add it to g
	g.SearchStatuses = append(g.SearchStatuses, search)
}

// DeleteSearchAfterTimeout starts a search timer and delete the search request
func (g *Gossiper) DeleteSearchAfterTimeout(req u.SearchRequest) {
	// wait for the timeout
	time.Sleep(u.DefaultDuplicateSearchTime)
	g.SearchMutex.Lock()
	// look for the status to delete
	for i, s := range g.SearchStatuses {
		if u.SameSearch(s, req) {
			// remove the search status from the list
			g.SearchStatuses[i] = g.SearchStatuses[len(g.SearchStatuses)-1]
			g.SearchStatuses = g.SearchStatuses[:len(g.SearchStatuses)-1]
			break
		}
	}
	g.SearchMutex.Unlock()
}
