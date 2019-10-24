package gossiper

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// Gossiper : a gossiper
type Gossiper struct {
	Name         string                 // name of the gossiper
	GossipAddr   *net.UDPAddr           // address of the gossip port
	ClientAddr   *net.UDPAddr           // address of the client port (udp4)
	GossipConn   *net.UDPConn           // connection for gossip
	ClientConn   *net.UDPConn           // ui connection
	GUIPort      int                    // gui port
	Peers        []net.UDPAddr          // list of direct peers (neighbors)
	Mode         string                 // mode of the gossiper (simple / rumor)
	PendingACKs  *sync.Map              // map[u.AckIdentifier]u.AckValues
	WantList     *sync.Map              // map[string]uint32
	RumorHistory *sync.Map              // map[string][]u.HistoryMessage
	AntiEntropy  int                    // anti entropy timeout value
	NewMessages  *u.SyncNewMessages     // slice containing all rumors
	Routes       *sync.Map              // map[string]string string of udp4
	RTimer       int                    // routing timer
	PrivateMsg   []u.PrivateMessage     // slice containing all private messages
	Printer      *log.Logger            // printer avoid concurrent stdout write
	FileStructs  []u.FileStruct         // known files
	FileStatus   []*u.FileRequestStatus // status to file requests
}

// NewGossiper : creates a new gossiper with the given parameters
func NewGossiper(address, name, UIPort, GUIPort, peerList *string,
	simple bool, rtimer, antiE int) *Gossiper {

	// define gossip address and connection for the new gossiper
	gossAddr, err := net.ResolveUDPAddr("udp4", *address)
	PanicCheck(err)
	gossConn, err := net.ListenUDP("udp4", gossAddr)
	PanicCheck(err)

	// sanitize the uiport for new gossiper
	cliPort, err := strconv.Atoi(*UIPort)
	if err != nil || cliPort < 0 || cliPort > 65535 {
		log.Fatalf("Error: invalid port %s", *UIPort)
	}

	guiPort, err := strconv.Atoi(*GUIPort)
	if err != nil || cliPort < 0 || cliPort > 65535 {
		log.Fatalf("Error: invalid port %s", *UIPort)
	}

	// define client address and connection for the new gossiper
	cliAddr := &net.UDPAddr{
		IP:   gossAddr.IP,
		Port: cliPort,
	}
	cliConn, err := net.ListenUDP("udp4", cliAddr)
	PanicCheck(err)

	peers := u.ParsePeers(peerList)
	var mode string
	if simple {
		mode = u.SimpleModeStr
	} else {
		mode = u.RumorModeStr
	}
	var acks sync.Map
	var history sync.Map
	var status sync.Map
	var routes sync.Map
	var pm []u.PrivateMessage
	var fstruct []u.FileStruct
	var fstatus []*u.FileRequestStatus

	var newMessages []u.RumorMessage
	nm := u.SyncNewMessages{Messages: newMessages}

	if *name == "" {
		*name = "Gossiper"
	}
	status.Store(*name, uint32(1))
	printer := log.New(os.Stdout, "", 0)

	return &Gossiper{
		Name:         *name,
		GossipAddr:   gossAddr,
		ClientAddr:   cliAddr,
		GossipConn:   gossConn,
		ClientConn:   cliConn,
		Peers:        *peers,
		Mode:         mode,
		PendingACKs:  &acks,
		WantList:     &status,
		RumorHistory: &history,
		AntiEntropy:  antiE,
		NewMessages:  &nm,
		GUIPort:      guiPort,
		Routes:       &routes,
		RTimer:       rtimer,
		PrivateMsg:   pm,
		Printer:      printer,
		FileStructs:  fstruct,
		FileStatus:   fstatus,
	}
}

// Listen : listen for new messages from clients
func (g *Gossiper) Listen(udpConn *net.UDPConn) {
	buf := make([]byte, u.BufferSize)

	for {
		// read new message
		m, addr, err := udpConn.ReadFromUDP(buf)
		ErrorCheck(err)

		// copy the receive buffer to avoid that it is modified while being used
		tmp := make([]byte, m)
		copy(tmp, buf[:m])

		// whenever a new message arrives, start a new go routine to handle it
		if udpConn == g.GossipConn { // gossip
			go g.HandleGossip(tmp, addr)
		} else { // message from client
			go g.HandleMessage(tmp, addr)
		}
	}
}

// DoAntiEntropy : manage anti entropy
func (g *Gossiper) DoAntiEntropy() {
	// if anti entropy is 0 (or less) anti entropy disabled
	if g.AntiEntropy > 0 {
		for {
			// sleep for anti entropy value
			time.Sleep(time.Duration(g.AntiEntropy) * time.Second)

			if len(g.Peers) > 0 {
				// send status to random peer
				target := g.GetRandPeer()
				g.SendStatus(target)
			}
		}
	}
}

// DoRouteRumors sends route rumors every rtime seconds
func (g *Gossiper) DoRouteRumors() {
	if g.RTimer > 0 {
		for {
			// sleep for rtimer value
			time.Sleep(time.Duration(g.RTimer) * time.Second)

			// if any peers
			if len(g.Peers) > 0 {
				// send route rumor to random peer
				g.SendRouteRumor()
			}
		}
	}
}

// Run : runs a given gossiper
func (g *Gossiper) Run() {
	go g.StartServer()

	// starts a listener on ui and gossip ports
	go g.Listen(g.ClientConn)
	go g.Listen(g.GossipConn)

	// send the initial route rumor
	go g.SendRouteRumor()

	// if in rumor mode, do anti entropy and sends route rumors
	if g.Mode == u.RumorModeStr {
		go g.DoRouteRumors()
		go g.DoAntiEntropy()
	}
	// keep the program active
	for {

	}
}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name, UIPort, GUIPort, peerList *string,
	simple bool, rtimer, antiE int) {
	NewGossiper(address, name, UIPort, GUIPort, peerList, simple,
		rtimer, antiE).Run()
}
