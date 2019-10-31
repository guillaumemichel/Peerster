package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	r "math/rand"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/dedis/protobuf"
)

// ProtobufMessage : encapsulate a client message using protobuf
func ProtobufMessage(msg *Message) []byte {
	// serializing the packet to send
	bytesToSend, err := protobuf.Encode(msg)
	if err != nil {
		log.Panic("Error: couldn't serialize message")
	}
	return bytesToSend
}

// UnprotobufMessage : decapsulate a gossip message using protobuf
func UnprotobufMessage(packet []byte) (*Message, bool) {
	rcvMsg := Message{}

	err := protobuf.Decode(packet, &rcvMsg)
	ok := true
	if err != nil {
		log.Println(err)
		ok = false
	}
	return &rcvMsg, ok
}

// ProtobufGossip : encapsulates a gossip message using protobuf
func ProtobufGossip(msg *GossipPacket) []byte {

	// serializing the packet to send
	bytesToSend, err := protobuf.Encode(msg)
	if err != nil {
		log.Panic("Error: couldn't serialize message")
	}
	return bytesToSend
}

// UnprotobufGossip : decapsulate a gossip message using protobuf
func UnprotobufGossip(packet []byte) (*GossipPacket, bool) {
	rcvMsg := GossipPacket{}

	err := protobuf.Decode(packet, &rcvMsg)
	ok := true
	if err != nil {
		log.Println(err)
		ok = false
	}
	return &rcvMsg, ok
}

// ParsePeers : infe
func ParsePeers(peerList *string) *[]net.UDPAddr {

	if *peerList == "" {
		return &[]net.UDPAddr{}
	}
	// split up the different addresses
	peersStr := strings.Split(*peerList, ",")

	// TODO: sanitize ip addresses

	var addrList []net.UDPAddr
	for _, v := range peersStr {
		// split up the ip from the port
		addr := strings.Split(v, ":")

		// parse ip address
		ip := net.ParseIP(addr[0])
		if ip == nil {
			log.Fatalf("Error: invalid IP address %s", addr[0])
		}

		// parse port
		port, err := strconv.Atoi(addr[1])
		if err != nil || port < 0 || port > 65535 {
			log.Fatalf("Error: invalid port %s", addr[1])
		}

		// add address to address List
		addrList = append(addrList, net.UDPAddr{
			IP:   ip,
			Port: port,
		})
	}

	return &addrList
}

// EqualAddr : compares the given UDPAddr, returns false if they are different
// and true otherwise
func EqualAddr(addr1, addr2 *net.UDPAddr) bool {
	if addr1 == nil || addr2 == nil {
		return false
	}
	return addr1.String() == addr2.String()
	/* return (bytes.Equal(addr1.IP, addr2.IP) && addr1.Port == addr2.Port &&
	addr1.Zone == addr2.Zone)*/
}

// GetRand : returns a "fake" random number (the same at all executions)
func GetRand(n int) int {
	return r.Intn(n)
}

// GetRealRand : returns a really random number generated with crypto package
func GetRealRand(n int) int {
	result, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(result.Int64())
	//return GetRand(n)
}

// TestMessageType : test if the packet only contains a message type and prints
// an error and returns false if it is not the case
func TestMessageType(p *GossipPacket) bool {
	// counter of pointers that are set in gossip packet
	n := 0
	if p.Simple != nil {
		n++
	}
	if p.Rumor != nil {
		n++
	}
	if p.Status != nil {
		n++
	}
	if p.Private != nil {
		n++
	}
	if p.DataReply != nil {
		n++
	}
	if p.DataRequest != nil {
		n++
	}
	if n > 1 {
		fmt.Println("Error: the received GossipPacket contains multiple",
			"messages")
		return false
	} else if n < 1 {
		fmt.Println("Error: the received GossipPacket don't contain any " +
			"message")
		return false
	}
	return true
}

// GetACKIdentifierSend : return the string corresponding to the ACK identifier
// of the given rumor when sending a message
func GetACKIdentifierSend(rumor *RumorMessage, dest *string) *AckIdentifier {
	identifer := AckIdentifier{
		Peer:   *dest,
		Origin: rumor.Origin,
		ID:     rumor.ID + 1,
	}
	return &identifer
}

// GetACKIdentifierReceive : return the string corresponding to the ACK
// identifier of the given ack when receiving a message
func GetACKIdentifierReceive(nextID uint32, rumorOrigin,
	sender *string) *AckIdentifier {
	identifier := AckIdentifier{
		Peer:   *sender,
		Origin: *rumorOrigin,
		ID:     nextID,
	}
	return &identifier
}

// SyncMapCount : count the number of elements in a sync map
func SyncMapCount(sm *sync.Map) uint32 {
	count := uint32(0)
	f := func(k, v interface{}) bool {
		count++
		return true
	}
	sm.Range(f)
	return count
}

// CheckFilename check if the filename is correct and return an error if not
func CheckFilename(name string) error {
	for _, c := range name {
		if byte(c) == 0 || c == '/' || c == '\\' || c == ':' {
			return errors.New("invalid filename given")
		}
	}
	return nil
}

/*
// RemoveAddrFromPeers : remove the given address from the array of addresses
func RemoveAddrFromPeers(peers *[]net.UDPAddr, addr *net.UDPAddr) *[]net.UDPAddr {

	for i, v := range *peers {
		if EqualAddr(&v, addr) && i == len(*peers)-1 {
			var toReturn []net.UDPAddr
			if i == len(*peers)-1 {
				toReturn = append((*peers)[:i])
			} else {
				toReturn = append((*peers)[:i], (*peers)[i+1:]...)
			}
			return &toReturn
		}
	}
	fmt.Println("Warning: couldn't remove address from peers")
	return peers
}
*/
