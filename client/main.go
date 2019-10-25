package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	u "github.com/guillaumemichel/Peerster/utils"
)

// BadArgument print bad argument error message and exit
func BadArgument() {
	fmt.Println("ERROR (Bad argument combination)")
	os.Exit(1)
}

// BadRequest prints bad request error message and exit
func BadRequest() {
	fmt.Println("ERROR (Unable to decode hex hash)")
	os.Exit(1)
}

func main() {
	// treating the flags
	UIPort := flag.String("UIPort", u.DefaultUIPort, "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; "+
		"can be omitted")
	file := flag.String("file", "", "file to be indexed by the gossiper")
	req := flag.String("request", "",
		"request a chunk or metafile of this hash")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// parse port
	port, err := strconv.Atoi(*UIPort)
	if err != nil || port < 0 || port > 65535 {
		log.Fatalf("Error: invalid port %s", *UIPort)
	}
	// parse destination
	ip := net.ParseIP(u.LocalhostAddr)

	// creating destination address
	address := net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	var message u.Message

	bText := *msg != ""
	bDest := dest != nil && *dest != ""
	bFile := file != nil && *file != ""
	bReq := req != nil && *req != ""

	// simple message or rumor message
	if bText && !bDest && !bFile && !bReq {
		// create the packet to send
		message = u.Message{
			Text:        *msg,
			Destination: nil,
			File:        nil,
			Request:     nil,
		}
	} else if bText && bDest && !bFile && !bReq {
		// private message
		// create the packet to send
		message = u.Message{
			Text:        *msg,
			Destination: dest,
			File:        nil,
			Request:     nil,
		}

	} else if !bText && bDest && bFile && !bReq {
		// sending file
		// create the send file message
		message = u.Message{
			Destination: dest,
			File:        file,
			Text:        "",
			Request:     nil,
		}
	} else if !bText && bDest && bFile && bReq {
		// requesting file
		// cast string request to []byte
		hashByte, err := hex.DecodeString(*req)
		if err != nil {
			BadRequest()
		}
		// create the request
		message = u.Message{
			Destination: dest,
			File:        file,
			Request:     &hashByte,
		}
	} else {
		BadArgument()
	}

	// creating upd connection
	udpConn, err := net.DialUDP("udp4", nil, &address)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// serializing the packet to send
	bytesToSend := u.ProtobufMessage(&message)

	// sending the packet over udp
	_, err = udpConn.Write(bytesToSend)
	if err != nil {
		fmt.Println(err)
	}
}
