package gossiper

// SendNewSearchReq creates and sends a new file search request
func (g *Gossiper) SendNewSearchReq(budget uint64, keywords []string) {
	req := u.SearchRequest{
		Origin: g.Name,
		Budget: budget,
		Keywords: keywords,
	}
}
