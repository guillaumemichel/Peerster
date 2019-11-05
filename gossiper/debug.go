package gossiper

import (
	"fmt"
	"log"

	u "github.com/guillaumemichel/Peerster/utils"
)

// ErrorCheck : check for non critical error, and logs the result
func ErrorCheck(err error) {
	if err != nil {
		log.Println(err)
	}
}

// PanicCheck : check for panic level errors, and logs the result
func PanicCheck(err error) {
	if err != nil {
		log.Panic(err)
	}
}

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

// PrintWantlist : prints the wantlist of the Gossiper
func (g *Gossiper) PrintWantlist() {
	f := func(k, v interface{}) bool {
		fmt.Printf("%s : %d\n", k.(string), v.(uint32))
		return true
	}
	fmt.Println("\nDEBUG: Printing wantlist")
	//g.WantListMutex.Lock()
	g.WantList.Range(f)
	//g.WantListMutex.Unlock()
	fmt.Println()
}
