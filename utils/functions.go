package utils

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/DeDiS/protobuf"
)

// ProtobufMessage : encapsulates a message using protobuf
func ProtobufMessage(msg *GossipPacket) []byte {

	// serializing the packet to send
	bytesToSend, err := protobuf.Encode(msg)
	if err != nil {
		log.Panic("Error: couldn't serialize message")
	}
	return bytesToSend
}

// UnprotobufMessage : decapsulate a message using protobuf
func UnprotobufMessage(packet []byte) (*GossipPacket, bool) {
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
