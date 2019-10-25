package gossiper

import (
	u "github.com/guillaumemichel/Peerster/utils"
)

// UpdateRoute updates the route to the given origin with the given nextHop for
// the gossiper
func (g *Gossiper) UpdateRoute(rumor u.RumorMessage, nextHop string) {
	origin := rumor.Origin
	// test if the message has an higher ID than the one from which the route
	// is already stored
	if rumor.ID > g.GetLastIDFromOrigin(origin) {
		// load the current route (if any)
		//v, ok := g.Routes.Load(origin)
		// if the route exists and is different of the given one, update it
		//if !ok || v.(string) != nextHop {
		// update route and print message
		g.Routes.Store(origin, nextHop)
		g.PrintUpdateRoute(origin, nextHop)
		//}
	}
}
