package gossiper

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	u "github.com/guillaumemichel/Peerster/utils"
)

// Gossiper : a gossiper
type Gossiper struct {
	Name       string       // name of the gossiper
	GossipAddr *net.UDPAddr // address of the gossip port
	ClientAddr *net.UDPAddr // address of the client port (udp4)
	GossipConn *net.UDPConn // connection for gossip
	ClientConn *net.UDPConn // ui connection
	GUIPort    int          // gui port

	Peers []net.UDPAddr // list of direct peers (neighbors)
	//PeerMutex *sync.Mutex

	Mode string // mode of the gossiper (simple / rumor)

	PendingACKs *sync.Map // map[u.AckIdentifier]u.AckValues
	//ACKMutex    *sync.Mutex

	WantList *sync.Map // map[string]uint32
	//WantListMutex *sync.Mutex

	RumorHistory *sync.Map // map[string][]u.HistoryMessage
	//HistoryMutex *sync.Mutex

	AntiEntropy int                // anti entropy timeout value
	NewMessages *u.SyncNewMessages // slice containing all rumors

	Routes     *sync.Map   // map[string]string string of udp4
	RouteMutex *sync.Mutex // mutex for the routes map

	RTimer      int                    // routing timer
	PrivateMsg  []u.PrivateMessage     // slice containing all private messages
	Printer     *log.Logger            // printer avoid concurrent stdout write
	FileStructs []u.FileStruct         // known files
	FileStatus  []*u.FileRequestStatus // status to file requests
	Chunks      []u.FileChunk          // list of chunks that the gossiper has
	ChunkLock   *sync.Mutex

	SearchStatuses []u.SearchStatus // list of current searches
	SearchMutex    *sync.Mutex

	SearchChans   map[*[]string]chan u.SearchReply
	SearchResults []u.SearchFile

	Hw3ex2 bool // hw3ex2 mode on or off
	Hw3ex4 bool // hw3ex4 mode on or off
	N      int  // number of connected peers
	AckAll bool // ack every message irrespective of ID

	StubbornTimeout time.Duration     // stubborn timeout
	BlockStatuses   map[string]uint32 // vector clock of blocks origin -> ID
	Round           uint32            // current round

	BlockChans map[uint32]*chan u.TLCAck // tlc ack channels
	BlockRumor map[string]map[string]map[uint32]*chan bool
	// addr -> origin -> ID -> chan

	LogLvl string // log level of the peerster
}

// NewGossiper : creates a new gossiper with the given parameters
func NewGossiper(address, name, UIPort, GUIPort, peerList *string,
	simple, hw3ex2, hw3ex4, ackAll bool, rtimer, antiE, stubbornTimeout, n int,
	loglvl string) *Gossiper {

	// define gossip address and connection for the new gossiper
	gossAddr, err := net.ResolveUDPAddr("udp4", *address)
	PanicCheck(err)
	gossConn, err := net.ListenUDP("udp4", gossAddr)
	PanicCheck(err)

	// sanitize the uiport for new gossiper
	cliPort, err := strconv.Atoi(*UIPort)
	if err != nil || cliPort < 0 || cliPort > 65535 {
		log.Fatalln(err, cliPort)
		log.Fatalf("Error: invalid port %s", *UIPort)
	}

	guiPort, err := strconv.Atoi(*GUIPort)
	if err != nil || cliPort < 0 || cliPort > 65535 {
		log.Fatalln(err, guiPort)
		log.Fatalf("Error: invalid port %s", *UIPort)
	}

	// define client address and connection for the new gossiper
	cliAddr := &net.UDPAddr{
		IP:   gossAddr.IP,
		Port: cliPort,
	}
	cliConn, err := net.ListenUDP("udp4", cliAddr)
	PanicCheck(err)

	peers := *u.ParsePeers(peerList)

	// remove the gossipAddr of g if it is in the host list
	for i, p := range peers {
		if p.String() == gossAddr.String() {
			// remove p
			if i == len(peers)-1 {
				peers = peers[:i]
			} else {
				peers = append(peers[:i], peers[i+1:]...)
			}
		}
	}

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
	var chunks []u.FileChunk
	var fstatus []*u.FileRequestStatus
	var searchs []u.SearchStatus
	/*
		fstructs := f.LoadSharedFiles()
	*/
	fstructs := make([]u.FileStruct, 0)
	schan := make(map[*[]string]chan u.SearchReply)
	brum := make(map[string]map[string]map[uint32]*chan bool)
	/*
		statuses := make([]u.FileRequestStatus, 0)
		fstatus := u.FileStatusList{List: statuses}
	*/
	bChans := make(map[uint32]*chan u.TLCAck)

	var newMessages []u.RumorMessage
	nm := u.SyncNewMessages{Messages: newMessages}

	if *name == "" {
		*name = u.DefaultGossiperName
	}

	bStatuses := make(map[string]uint32)
	bStatuses[*name] = 0 // block vector clock starts at 0

	status.Store(*name, uint32(1))
	printer := log.New(os.Stdout, "", 0)

	return &Gossiper{
		Name:       *name,
		GossipAddr: gossAddr,
		ClientAddr: cliAddr,
		GossipConn: gossConn,
		ClientConn: cliConn,
		Peers:      peers,
		//PeerMutex:     &sync.Mutex{},
		Mode:        mode,
		PendingACKs: &acks,
		//ACKMutex:      &sync.Mutex{},
		WantList: &status,
		//WantListMutex: &sync.Mutex{},
		RumorHistory: &history,
		//HistoryMutex:  &sync.Mutex{},
		AntiEntropy:    antiE,
		NewMessages:    &nm,
		GUIPort:        guiPort,
		Routes:         &routes,
		RouteMutex:     &sync.Mutex{},
		RTimer:         rtimer,
		PrivateMsg:     pm,
		Printer:        printer,
		FileStructs:    fstructs,
		FileStatus:     fstatus,
		Chunks:         chunks,
		ChunkLock:      &sync.Mutex{},
		SearchStatuses: searchs,
		SearchMutex:    &sync.Mutex{},
		SearchChans:    schan,
		Hw3ex2:         hw3ex2,
		Hw3ex4:         hw3ex4,
		N:              n,
		AckAll:         ackAll,
		BlockStatuses:  bStatuses,
		Round:          0,
		BlockChans:     bChans,
		BlockRumor:     brum,
		LogLvl:         loglvl,
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
			//g.PeerMutex.Lock()
			l := len(g.Peers)
			//g.PeerMutex.Unlock()
			if l > 0 {
				// send status to random peer
				target := g.GetRandPeer()
				g.SendStatus(target)
			}

			// sleep for anti entropy value
			time.Sleep(time.Duration(g.AntiEntropy) * time.Second)
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

	// starts a listener on ui and gossip ports
	go g.Listen(g.ClientConn)
	go g.Listen(g.GossipConn)

	// if in rumor mode, do anti entropy and sends route rumors
	if g.Mode == u.RumorModeStr {
		// start server
		go g.StartServer()
		go g.DoRouteRumors()
		go g.DoAntiEntropy()
	}

	// keep the program active until ctrl+c is pressed
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)
}

// StartNewGossiper : Creates and starts a new gossiper
func StartNewGossiper(address, name, UIPort, GUIPort, peerList *string,
	simple, hw3ex2, hw3ex4, ackAll bool, rtimer, antiE, stubbornTimeout,
	n int, loglvl string) {

	NewGossiper(address, name, UIPort, GUIPort, peerList, simple, hw3ex2,
		hw3ex4, ackAll, rtimer, antiE, stubbornTimeout, n, loglvl).Run()
}
