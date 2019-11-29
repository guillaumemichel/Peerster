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
	g.WantListMutex.Lock()
	str := "\nDEBUG: Printing wantlist\n"
	for k, v := range g.WantList {
		str += fmt.Sprintf("%s : %d, ", k, v)
	}
	g.WantListMutex.Unlock()
	str += "\n\n"
	g.Printer.Println(str)
}
