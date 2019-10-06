package main

import (
	"flag"
	"fmt"
	"os"

	g "github.com/guillaumemichel/Peerster/gossiper"
)

func main() {

	UIPort := flag.String("UIPort", "8080", "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000",
		"ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peersInput := flag.String("peers", "",
		"comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false,
		"run gossiper in simple broadcast mode")
	//bufferSize := flag.Int("buffer-size", 1024, "buffer size of the udp socket")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	if !*simple {
		fmt.Println("Sorry, only simple messages implemented yet :)")
		os.Exit(0)
	}

	g.StartNewGossiper(gossipAddr, name, UIPort, peersInput)
}
