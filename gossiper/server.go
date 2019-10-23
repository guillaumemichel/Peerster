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

	getLatestRumorMessagesHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s := r.FormValue("number")
			n, _ := strconv.Atoi(s)

			msgList := g.GetNewMessages(n)
			if msgList != nil {
				msgListJSON, _ := json.Marshal(msgList)
				// error handling, etc...
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(msgListJSON)
			}
		}
	}

	getPeers := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s := r.FormValue("number")
			n, _ := strconv.Atoi(s)

			list := g.GetNewPeers(n)
			if list != nil {
				msgListJSON, _ := json.Marshal(list)
				// error handling, etc...
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(msgListJSON)
			}
		}
	}

	sendMsg := func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			msg := r.FormValue("message")
			dst := r.FormValue("dest")
			g.SendMessage(msg, dst)
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

	destinationHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// load last dest received
			s := r.FormValue("dest_count")
			n, _ := strconv.Atoi(s)
			// get the dest added after last
			dests := g.GetDestinations(uint32(n))
			// if there are some new dests
			if dests != nil {
				peersJSON, _ := json.Marshal(dests)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(peersJSON)
			}
		}
	}

	pmHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// load last dest received
			s := r.FormValue("last_pm")
			n, _ := strconv.Atoi(s)
			// get the dest added after last
			pms := g.GetPrivateMessage(n)
			// if there are some new dests
			if pms != nil {
				pmJSON, _ := json.Marshal(pms)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(pmJSON)
			}
		}
	}

	http.Handle("/", http.FileServer(http.Dir("gui/html/")))
	http.HandleFunc("/id", getGossiperID)
	http.HandleFunc("/node", getPeers)
	http.HandleFunc("/message", getLatestRumorMessagesHandler)
	http.HandleFunc("/send", sendMsg)
	http.HandleFunc("/newpeer", addPeer)
	http.HandleFunc("/updatedest", destinationHandler)
	http.HandleFunc("/pm", pmHandler)

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
