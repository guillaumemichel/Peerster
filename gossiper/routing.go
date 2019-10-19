package gossiper

import (
	"fmt"

	u "github.com/guillaumemichel/Peerster/utils"
)

// UpdateRoute updates the route to the given origin with the given nextHop for
// the gossiper
func (g *Gossiper) UpdateRoute(rumor u.RumorMessage, nextHop string) {
	origin := rumor.Origin
	fmt.Println(g.GetLastIDFromOrigin(origin))
	if rumor.ID > g.GetLastIDFromOrigin(origin) {
		v, ok := g.Routes.Load(origin)
		if !ok || v.(string) != nextHop {
			g.Routes.Store(origin, nextHop)
			g.PrintUpdateRoute(origin, nextHop)
		}
	}
}
