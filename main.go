package main

import (
	"flag"
	"fmt"
	"os"

	f "github.com/guillaumemichel/Peerster/files"
	g "github.com/guillaumemichel/Peerster/gossiper"
	u "github.com/guillaumemichel/Peerster/utils"
)

func main() {

	// create the directories shared folder and downloads
	f.CheckSharedDir()
	f.CheckDownloadDir()

	hoplim := int(u.DefaultHopLimit)

	// flags of the peerster command
	UIPort := flag.String("UIPort", u.DefaultUIPort, "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000",
		"ip:port for the gossiper")
	name := flag.String("name", "258066", "name of the gossiper")
	peersInput := flag.String("peers", "",
		"comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false,
		"run gossiper in simple broadcast mode")
	antiE := flag.Int("antiEntropy", u.AntiEntropyDefault,
		"Use the given timeout in seconds for anti-entropy. If "+
			"the flag is absent, the default anti-entropy duration is 10 "+
			"seconds.")
	GUIPort := flag.String("GUIPort", u.DefaultGUIPort,
		"port for the GUI client")
	rtimer := flag.Int("rtimer", u.RTimerDefault, "Timeout in seconds to "+
		"send route rumors. 0 (default) means disable sending route rumors.")
	stubbornTimeout := flag.Int("stubbornTimeout", u.StubbornTimeoutDefault,
		"timeout to resend TLC message if not acked by a majority of peers")
	loglvl := flag.String("debug", "111",
		"debug flags that correspond to HW1-2-3")
	hw3ex2Flag := flag.Bool("hw3ex2", false, "set to true to publish blocks"+
		" when indexing files")
	hw3ex4Flag := flag.Bool("hw3ex4", false, "set to true enable QSC")
	n := flag.Int("N", u.DefaultPeerNumber, "number of connected peers")
	ackAll := flag.Bool("ackAll", true, "ack every message irrespective"+
		" of its ID")
	ackHopLimit := flag.Int("hopLimit", hoplim,
		"hop limit for TLC ack messages")

	flag.Parse()
	// help message
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// start new gossiper
	g.StartNewGossiper(gossipAddr, name, UIPort, GUIPort, peersInput,
		*simple, *hw3ex2Flag, *hw3ex4Flag, *ackAll, *rtimer, *antiE,
		*stubbornTimeout, *n, *ackHopLimit, *loglvl)
}
