package gossiper

// UpdateRoute updates the route to the given origin with the given nextHop for
// the gossiper
func (g *Gossiper) UpdateRoute(origin, nextHop string) {
	g.Routes.Store(origin, nextHop)
	g.PrintUpdateRoute(origin, nextHop)
}
