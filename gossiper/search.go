package gossiper

import (
	"encoding/hex"
	"net"
	"sort"
	"strings"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ManageSearch manage a search
func (g *Gossiper) ManageSearch(initialBudget *uint64, keywords []string) {
	g.SearchChans[&keywords] = make(chan u.SearchReply)

	endChan := make(chan bool)
	doubleBudget := initialBudget == nil

	if g.ShouldPrint(logHW3, 2) {
		var bPrint uint64
		if doubleBudget {
			bPrint = u.DefaultSearchBudget
		} else {
			bPrint = *initialBudget
		}

		g.Printer.Printf("Got a file search request from client \n  "+
			"Budget: %d\n  Keywords: %s\n\n", bPrint,
			u.FlattenKeywords(keywords))
	}

	// function that resend the request every second until receiving a Signal
	// from endChan
	go func() {
		var b uint64
		if doubleBudget {
			b = u.DefaultSearchBudget
		} else {
			b = *initialBudget
		}
		timeout := make(chan bool)
		for {
			// send request
			g.SendNewSearchReq(b, keywords)

			// timeout
			go func() {
				// timeout value defined in utils/constants.go
				time.Sleep(time.Duration(u.SearchPeriod))
				timeout <- true
			}()

			select {
			case <-endChan:
				// end
				return
			case <-timeout:
				// doubling budget (except if already near max)
				if doubleBudget && b < u.MaxSearchBudget/2 {
					b *= 2
					if g.ShouldPrint(logHW3, 2) {
						g.Printer.Println("Doubling search budget")
					}

				} else {
					if g.ShouldPrint(logHW3, 2) {
						g.Printer.Println("Search failed!")
					}
					return
				}
			}
		}
	}()

	files := make([]u.SearchFile, 0)
	var h u.ShaHash
	doneCount := 0

	// exit the loop when threshold condition is met
	for doneCount < u.MatchThreshold {
		select {
		// waits for search reply corresponding to keywords
		case rep := <-g.SearchChans[&keywords]:
			// iterates over the results of the reply
			for _, v := range rep.Results {
				copy(h[:], v.MetafileHash)

				// check if the result matches with my keywords
				myRes := false
				for _, w := range keywords {
					if strings.Contains(v.FileName, w) {
						myRes = true
						break
					}
				}

				if myRes {
					// iterate over the search files we already have
					found := false
					for _, sf := range files {
						// if file already is in files (same metahash)
						if sf.MetafileHash == h {
							// update host's chunks
							for _, i := range v.ChunkMap {
								sf.Chunks[i-1][rep.Origin] = true
							}
							// check if file is complete
							g.CheckSearchFileComplete(&sf)
							found = true
						}
					}
					// if file doesn't exist yet in files, create it
					if !found {
						cmap := make(map[uint64]map[string]bool)

						for i := uint64(0); i < v.ChunkCount; i++ {
							cmap[i] = make(map[string]bool)
						}

						f := u.SearchFile{
							Name:         v.FileName,
							MetafileHash: h,
							Chunks:       cmap,
							NChunks:      v.ChunkCount,
							Complete:     false,
						}
						// update host
						for _, i := range v.ChunkMap {
							f.Chunks[i-1][rep.Origin] = true
						}
						// check if file is complete
						g.CheckSearchFileComplete(&f)
						// append the freshly created file to files
						files = append(files, f)
					}
				}
			}
		}
		// check how many files are complete
		doneCount = 0
		for _, sf := range files {
			if sf.Complete {
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("We got", sf.Name)
				}
				doneCount++
			}
		}
	}
	endChan <- true
	// add the results to the gossiper
	g.SearchResults = append(g.SearchResults, files...)
	g.PrintSearchFinished()
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
	g.HandleSearchReq(req, nil)
}

// HandleSearchReq handles a search request
func (g *Gossiper) HandleSearchReq(req u.SearchRequest, sender *net.UDPAddr) {

	// register request
	g.SearchMutex.Lock()
	// check if search request is duplicate
	if g.NewSearch(req) {
		// if not register it
		g.AddSearch(req)
		g.SearchMutex.Unlock()

		// and deletes it after a timeout
		go g.DeleteSearchAfterTimeout(req)
	} else {
		g.SearchMutex.Unlock()
		return
	}

	if g.ShouldPrint(logHW3, 2) {
		g.Printer.Printf("Got search request from %s\n  Budget: %d\n  "+
			"Keywords: %s\n  Peers: %s\n\n", req.Origin, req.Budget,
			u.FlattenKeywords(req.Keywords), g.PeersToString())

	}

	// search for file locally
	if sender != nil {
		g.SearchForFile(req.Origin, req.Keywords)
	}
	// substract 1 to the budget
	req.Budget--

	if len(g.Peers) == 0 {
		return
	}

	if req.Budget > 0 {
		quotient := int(req.Budget) / len(g.Peers)
		rest := int(req.Budget) % len(g.Peers)

		nPeers := len(g.Peers)
		peers := g.Peers
		if sender != nil {
			// remove sending peer
			nPeers--
			peers = u.RemoveAddrFromPeers(peers, *sender)
		}
		randPeers := u.ChooseCRandomAmongN(rest, nPeers)
		if randPeers == nil {
			return
		}

		for i, p := range peers {
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
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Sending request to", p.String(),
						"with budget", newReq.Budget)
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
	g.ChunkLock.Lock()
	if g.ShouldPrint(logHW3, 2) {
		g.Printer.Println("Chunk number:", len(g.Chunks))
	}
	for _, c := range g.Chunks {
		// iterate over all keywords
		for _, w := range keywords {
			// if the name of the chunk correspond to a keyword
			if strings.Contains(c.File.Name, w) {
				if g.ShouldPrint(logHW3, 2) {
					g.Printer.Println("Match to request! File:", c.File.Name)
				}
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
	g.ChunkLock.Unlock()

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
		/*
			// increase the index number as they should start at 1
			for i := range cmap {
				cmap[i]++
			}
		*/

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
	if dest == g.Name {
		// if lookup on the sender
		g.TreatMySearchReply(rep)
	} else {
		g.HandleSearchReply(rep)
	}
}

// TreatMySearchReply handle the search relpy that are for me
func (g *Gossiper) TreatMySearchReply(rep u.SearchReply) {
	// the reply is for me
	for _, r := range rep.Results {
		//string metahash
		hash := hex.EncodeToString(r.MetafileHash)
		// print found match message
		g.PrintFoundMatch(r.FileName, rep.Origin, hash, r.ChunkMap)
	}
	// send the reply in all channels
	for _, c := range g.SearchChans {
		c <- rep
	}
}

// HandleSearchReply handles search replies
func (g *Gossiper) HandleSearchReply(rep u.SearchReply) {

	if g.ShouldPrint(logHW3, 3) {
		g.Printer.Printf("Destination: %s   My name: %s\n\n",
			rep.Destination, g.Name)
	}
	if rep.Destination == g.Name {
		g.TreatMySearchReply(rep)
		return
	}

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
