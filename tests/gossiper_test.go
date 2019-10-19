package tests

import (
	"net"
	"os"
	"sync"
	"testing"

	"github.com/guillaumemichel/Peerster/gossiper"
	u "github.com/guillaumemichel/Peerster/utils"
)

// PanicCheck checks for critical errors
func PanicCheck(err error) {
	if err != nil {
		os.Exit(1)
	}
}

// GetGossiper1 returns a sample gossiper
func GetGossiper1() gossiper.Gossiper {
	name := "Gossiper1"
	addr := "127.0.0.1:5100"
	cliPort := 8090
	guiPort := 8090
	antiEntropy := 10
	peerList := "127.0.0.1:5101,127.0.0.1:5103"

	mode := u.RumorModeStr
	var acks sync.Map
	var history sync.Map
	var routes sync.Map
	var status sync.Map
	status.Store(name, uint32(1))

	var newMessages []u.RumorMessage
	nm := u.SyncNewMessages{Messages: newMessages}

	gossAddr, err := net.ResolveUDPAddr("udp4", addr)
	PanicCheck(err)
	gossConn, err := net.ListenUDP("udp4", gossAddr)
	PanicCheck(err)

	cliAddr := &net.UDPAddr{
		IP:   gossAddr.IP,
		Port: cliPort,
	}
	cliConn, err := net.ListenUDP("udp4", cliAddr)
	PanicCheck(err)

	peers := u.ParsePeers(&peerList)

	return gossiper.Gossiper{
		Name:         name,
		GossipAddr:   gossAddr,
		ClientAddr:   cliAddr,
		GossipConn:   gossConn,
		ClientConn:   cliConn,
		Peers:        *peers,
		BufSize:      u.BufferSize,
		Mode:         mode,
		PendingACKs:  &acks,
		WantList:     &status,
		RumorHistory: &history,
		AntiEntropy:  antiEntropy,
		NewMessages:  &nm,
		GUIPort:      guiPort,
		Routes:       &routes,
	}

}

// TestUpdateRoute test
func TestUpdateRoute(t *testing.T) {
	goss := GetGossiper1()
	g := &goss

	n1 := "127.0.0.1:5101"

	rumor1 := u.RumorMessage{
		Origin: "A",
		ID:     1,
		Text:   "blabla",
	}

	rumor2 := u.RumorMessage{
		Origin: "B",
		ID:     1,
		Text:   "hehe",
	}

	rumor3 := u.RumorMessage{
		Origin: "A",
		ID:     2,
		Text:   "haha",
	}

	if u.SyncMapCount(g.Routes) != 0 {
		t.Errorf("Route map not empty when starting\n")
	}
	g.UpdateRoute(rumor1, n1)
	g.WriteRumorToHistory(rumor1)
	if u.SyncMapCount(g.Routes) != 1 {
		t.Errorf("New route not added (A)\n")
	}
	g.UpdateRoute(rumor1, "127.0.0.1:5103")
	g.WriteRumorToHistory(rumor1)
	v, ok := g.Routes.Load("A")
	if u.SyncMapCount(g.Routes) != 1 || !ok || v.(string) != n1 {
		t.Errorf("New route updated while it should have been\n")
	}
	g.UpdateRoute(rumor2, n1)
	g.WriteRumorToHistory(rumor2)
	if u.SyncMapCount(g.Routes) != 2 {
		t.Errorf("New route not added (B)\n")
	}
	g.UpdateRoute(rumor3, "127.0.0.1:5103")
	g.WriteRumorToHistory(rumor3)
	if u.SyncMapCount(g.Routes) != 2 {
		t.Errorf("New route not updated (A)\n")
	}

}
