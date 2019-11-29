package gossiper

import (
	u "github.com/guillaumemichel/Peerster/utils"
)

// UpdateRoute updates the route to the given origin with the given nextHop for
// the gossiper
func (g *Gossiper) UpdateRoute(rumor u.RumorMessage, nextHop string) {
	origin := rumor.Origin
	if origin == g.Name {
		return
	}
	// test if the message has an higher ID than the one from which the route
	// is already stored
	g.HistoryMutex.Lock()
	l := len(g.PacketHistory[origin])
	g.HistoryMutex.Unlock()
	if int(rumor.ID) > l {
		// load the current route (if any)
		//v, ok := g.Routes.Load(origin)
		// if the route exists and is different of the given one, update it
		//if !ok || v.(string) != nextHop {
		// update route and print message
		g.RouteMutex.Lock()
		g.Routes[origin] = nextHop
		g.RouteMutex.Unlock()
		if rumor.Text != "" {
			g.PrintUpdateRoute(origin, nextHop)
		}
		//}
	}
}
