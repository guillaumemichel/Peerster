package gossiper

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	u "github.com/guillaumemichel/Peerster/utils"
)

// StartServer : starts a server
func (g *Gossiper) StartServer() {
	// Somewhere, maybe even in a different package you create.

	getLatestRumorMessagesHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			msgList := g.GetNewMessages()
			msgListJSON, _ := json.Marshal(msgList)
			// error handling, etc...
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(msgListJSON)
		}
	}

	getPeers := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			list := g.GetPeers()
			msgListJSON, _ := json.Marshal(list)
			// error handling, etc...
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(msgListJSON)
		}
	}

	sendMsg := func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			msg := r.FormValue("message")
			g.SendMessage(msg)
		}
	}

	addPeer := func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			msg := r.FormValue("peer")
			addr, err := net.ResolveUDPAddr("udp4", msg)
			if err == nil {
				g.AddPeer(addr)
			}
		}
	}

	getGossiperID := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			id := g.GetPeerID()
			msgListJSON, _ := json.Marshal(id)
			// error handling, etc...
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(msgListJSON)
		}
	}

	http.Handle("/", http.FileServer(http.Dir("gui/html/")))
	http.HandleFunc("/id", getGossiperID)
	http.HandleFunc("/peers", getPeers)
	http.HandleFunc("/messages", getLatestRumorMessagesHandler)
	http.HandleFunc("/send", sendMsg)
	http.HandleFunc("/newpeer", addPeer)

	address := u.LocalhostAddr + ":" + strconv.Itoa(g.GUIPort)
	ok := true
	for ok {
		err := http.ListenAndServe(address, nil)
		if err != nil {
			//println("Error: cannot start GUI")
			ok = false
		}
	}

}
