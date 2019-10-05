package peers

import (
	"log"
	"net"
	"strconv"
	"strings"
)

// PeerList : peerlist that every peer will maintain about other known hosts
type PeerList struct {
	peers []net.UDPAddr
}

// ParsePeers : infe
func ParsePeers(peerList string) *PeerList {

	if peerList == "" {
		return &PeerList{}
	}
	// split up the different addresses
	peersStr := strings.Split(peerList, ",")

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

	return &PeerList{peers: addrList}
}
