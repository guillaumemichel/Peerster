package testing

import (
	"encoding/hex"
	"fmt"

	u "github.com/guillaumemichel/Peerster/utils"
)

// PrintDataRequestR d
func PrintDataRequestR(dreq u.DataRequest) {
	fmt.Println("\n--- DATA REQUEST RECEIVED ---")
	fmt.Println("Origin: " + dreq.Origin)
	fmt.Println("Destination: " + dreq.Destination)
	fmt.Println("Hop Limit:", dreq.HopLimit)
	fmt.Println("Hash:", hex.EncodeToString(dreq.HashValue))
	fmt.Println("------------------------------\n ")
}

// PrintDataReplyR .
func PrintDataReplyR(drep u.DataReply) {
	fmt.Println("\n--- DATA REQUEST RECEIVED ---")
	fmt.Println("Origin: " + drep.Origin)
	fmt.Println("Destination: " + drep.Destination)
	fmt.Println("Hop Limit:", drep.HopLimit)
	fmt.Println("Hash:", hex.EncodeToString(drep.HashValue))
	fmt.Println("Data:", hex.EncodeToString(drep.Data))
	fmt.Println("------------------------------\n ")

}
