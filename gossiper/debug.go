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

// DebugStatusPacket : dbg
func DebugStatusPacket(packet *[]byte) {
	gossip, ok := u.UnprotobufGossip(*packet)
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
