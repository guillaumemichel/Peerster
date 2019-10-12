package gossiper

import (
	"fmt"

	u "github.com/guillaumemichel/Peerster/utils"
)

// Debug : print debug message
func Debug(title, msg string) {
	fmt.Println()
	fmt.Println("Debug:")
	fmt.Println(title)
	fmt.Println(msg)
	fmt.Println()
}

// PrintWantlist : prints the wantlist of a gossiper
func PrintWantlist(g *Gossiper) {
	fmt.Println("\nDebug: wantlist")
	for k, v := range g.WantList {
		fmt.Printf("Name %s, NextID %d\n", k, v)
	}
}

// DebugStatusPacket : dbg
func DebugStatusPacket(packet *[]byte) {
	gossip, ok := u.UnprotobufGossip(packet)
	fmt.Println(gossip)
	fmt.Println(ok)
	if gossip.Status != nil {
		fmt.Println("DEBUG status\nOK:", ok)
		list := gossip.Status.Want
		for _, v := range list {
			fmt.Println(v.Identifier, v.NextID)
		}

	}
}
