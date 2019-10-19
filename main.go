package main

import (
	"flag"
	"fmt"
	"os"

	g "github.com/guillaumemichel/Peerster/gossiper"
	u "github.com/guillaumemichel/Peerster/utils"
)

func main() {

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

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	g.StartNewGossiper(gossipAddr, name, UIPort, GUIPort, peersInput,
		*simple, *rtimer, *antiE)
}
