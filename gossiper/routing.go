package gossiper

import (
	"fmt"
	"net"

	u "github.com/guillaumemichel/Peerster/utils"
)

// UpdateRoute updates the route to the given origin with the given nextHop for
// the gossiper
func (g *Gossiper) UpdateRoute(gp u.GossipPacket, nextHop string) {
	var origin string
	var id int

	if gp.Rumor != nil {
		origin = gp.Rumor.Origin
		id = int(gp.Rumor.ID)
	} else if gp.TLCMessage != nil {
		origin = gp.TLCMessage.Origin
		id = int(gp.TLCMessage.ID)
	} else {
		return
	}

	if origin == g.Name {
		return
	}
	// test if the message has an higher ID than the one from which the route
	// is already stored
	g.HistoryMutex.Lock()
	l := len(g.PacketHistory[origin])
	g.HistoryMutex.Unlock()
	if id > l {
		// load the current route (if any)
		//v, ok := g.Routes.Load(origin)
		// if the route exists and is different of the given one, update it
		//if !ok || v.(string) != nextHop {
		// update route and print message
		g.RouteMutex.Lock()
		g.Routes[origin] = nextHop
		g.RouteMutex.Unlock()

		if (gp.Rumor != nil && gp.Rumor.Text != "") || gp.TLCMessage != nil {
			g.PrintUpdateRoute(origin, nextHop)
		}
	}
}

// RoutePacket routes private messages, data requests and replies to next hop
func (g *Gossiper) RoutePacket(dst string, gp u.GossipPacket) {
	// load the next hop to destination
	g.RouteMutex.Lock()
	v, ok := g.Routes[dst]
	g.RouteMutex.Unlock()

	if !ok {
		g.Printer.Println("Searchreply:", gp.SearchReply)
		g.Printer.Println("No route to", dst)
		return
	}

	// resolve the udp address of the next hop
	addr, err := net.ResolveUDPAddr("udp4", v)
	if err != nil {
		fmt.Println(err)
		return
	}
	if g.ShouldPrint(logHW3, 3) {
		g.Printer.Println("Routing packet to", *addr)
	}
	g.SendPacketToNeighbor(*addr, gp)
}

// SendPacketToNeighbor send packet to known neighbor
func (g *Gossiper) SendPacketToNeighbor(addr net.UDPAddr,
	gp u.GossipPacket) {
	// protobuf the packet
	packet := u.ProtobufGossip(&gp)

	// send the packet
	g.GossipConn.WriteToUDP(packet, &addr)
}
