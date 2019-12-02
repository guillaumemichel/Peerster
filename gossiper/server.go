package gossiper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	u "github.com/guillaumemichel/Peerster/utils"
)

const (
	postreq = "POST"
	getreq  = "GET"
)

// StartServer : starts a server
func (g *Gossiper) StartServer() {

	downloadFile := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
			s := r.FormValue("number")
			n, _ := strconv.Atoi(s)
			fmt.Println(n)
			if len(g.GUISearchResults) > n {
				str := g.GUISearchResults[n]
				str = strings.Split(strings.Split(str, "=")[1], " ")[0]
				h, err := hex.DecodeString(str)
				if err != nil {
					g.Printer.Println(err)
					return
				}
				var hash u.ShaHash
				copy(hash[:], h)
				for _, r := range g.SearchResults {
					if r.MetafileHash == hash && r.Complete {
						g.HandleDownload(r.Name, h)
						break
					}
				}
			}
			// manage filedownload

			//g.SearchResults
		}
	}

	searchRes := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
			msgList := g.GUISearchResults
			if len(msgList) > 0 {
				msgListJSON, _ := json.Marshal(msgList)
				// error handling, etc...
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(msgListJSON)
			}
		}
	}

	getTLCRound := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
			round := g.GetTLCRound()
			msgListJSON, _ := json.Marshal(round)
			// error handling, etc...
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(msgListJSON)
		}
	}

	committedHistory := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
			s := r.FormValue("number")
			n, _ := strconv.Atoi(s)
			msgList := g.GetCommittedHistory(n)
			if msgList != nil {
				msgListJSON, _ := json.Marshal(msgList)
				// error handling, etc...
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(msgListJSON)
			}
		}
	}

	blockchainMessageHistory := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
			s := r.FormValue("number")
			n, _ := strconv.Atoi(s)
			msgList := g.GetBlockchainMessageHistory(n)
			if msgList != nil {
				msgListJSON, _ := json.Marshal(msgList)
				// error handling, etc...
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(msgListJSON)
			}
		}
	}

	sendSearchReq := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
			search := r.FormValue("search")
			keywords := u.SplitKeywords(search)
			n := r.FormValue("budget")
			nn, _ := strconv.Atoi(n)
			budget := uint64(nn)
			g.ManageSearch(&budget, keywords)
		}
	}

	getLatestRumorMessagesHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
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
		case getreq:
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
		case postreq:
			msg := r.FormValue("message")
			dst := r.FormValue("dest")
			g.SendMessage(msg, dst)
		}
	}

	addPeer := func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case postreq:
			msg := r.FormValue("peer")
			addr, err := net.ResolveUDPAddr("udp4", msg)
			if err == nil {
				g.AddPeer(addr)
			}
		}
	}

	getGossiperID := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case getreq:
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
		case getreq:
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
		case getreq:
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

	indexFile := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case postreq:
			fn := r.FormValue("filename")
			err := u.CheckFilename(fn)
			if err == nil {
				fn = filepath.Base(fn)
				g.IndexFile(fn)
			} else {
				g.Printer.Println(err)
			}
		}
	}

	requestFile := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case postreq:
			fn := r.FormValue("filename")
			hash := r.FormValue("hash")
			dest := r.FormValue("destination")

			err := u.CheckFilename(fn)
			if err == nil {
				h, err := hex.DecodeString(hash)
				if err != nil {
					g.Printer.Println(err)
					return
				}
				g.RequestFile(fn, dest, h)
			} else {
				g.Printer.Println(err)
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
	http.HandleFunc("/index", indexFile)
	http.HandleFunc("/requestfile", requestFile)
	http.HandleFunc("/search", sendSearchReq)
	http.HandleFunc("/searchresults", searchRes)
	http.HandleFunc("/blockchain", blockchainMessageHistory)
	http.HandleFunc("/committed_history", committedHistory)
	http.HandleFunc("/tlc", getTLCRound)
	http.HandleFunc("/download", downloadFile)

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
