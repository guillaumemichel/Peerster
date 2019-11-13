package gossiper

import (
	"fmt"

	u "github.com/guillaumemichel/Peerster/utils"
)

// SendNewSearchReq creates and sends a new file search request
func (g *Gossiper) SendNewSearchReq(budget uint64, keywords []string) {
	req := u.DataRequest{
		Origin: g.Name,
	}
	fmt.Println(req)
}
