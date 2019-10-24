package gossiper

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"

	u "github.com/guillaumemichel/Peerster/utils"
)

// CreateSimpleMessage : creates a gossip simple message from a string
func (g *Gossiper) CreateSimpleMessage(name, content string) u.SimpleMessage {
	return u.SimpleMessage{
		OriginalName:  name,
		RelayPeerAddr: g.GossipAddr.String(),
		Contents:      content,
	}
}

// CreateRumorMessage :creates a gossip rumor message from a string, and
// increases the rumorCount from the Gossiper
func (g *Gossiper) CreateRumorMessage(content string) u.RumorMessage {

	id, ok := g.WantList.Load(g.Name)
	if !ok {
		fmt.Println("Error: I don't know my name")
	}
	g.WantList.Store(g.Name, uint32(id.(uint32)+1))

	return u.RumorMessage{
		Origin: g.Name,
		ID:     id.(uint32),
		Text:   content,
	}
}

// CreateRouteMessage returns a route rumor message
func (g *Gossiper) CreateRouteMessage() u.RumorMessage {
	return g.CreateRumorMessage("")
}

// CreatePrivateMessage returns a private message with given text and
// destination
func (g *Gossiper) CreatePrivateMessage(text, dst string) u.PrivateMessage {
	return u.PrivateMessage{
		Origin:      g.Name,
		ID:          u.PrivateMessageID,
		Text:        text,
		Destination: dst,
		HopLimit:    u.DefaultHopLimit,
	}
}

// ReplaceRelayPeerSimple : replaces the relay peer of a simple message with its
// own address
func (g *Gossiper) ReplaceRelayPeerSimple(
	msg *u.SimpleMessage) *u.SimpleMessage {

	msg.RelayPeerAddr = g.GossipAddr.String()
	if msg.RelayPeerAddr == "<nil>" {
		log.Fatal("cannot replace relay peer address")
	}
	return msg
}

// BuildStatusPacket : build a status packet for g
func (g *Gossiper) BuildStatusPacket() u.StatusPacket {

	var want []u.PeerStatus

	f := func(k, v interface{}) bool {
		want = append(want, u.PeerStatus{Identifier: k.(string),
			NextID: v.(uint32)})
		return true
	}

	g.WantList.Range(f)
	sp := u.StatusPacket{Want: want}
	return sp
}

// SendMessage : sends a client message to the gossiper
func (g *Gossiper) SendMessage(text, dest string) {

	if text == "" {
		fmt.Println("Error: message required!")
		os.Exit(1)
	}

	// creating destination address
	address := g.ClientAddr

	// creating upd connection
	udpConn, err := net.DialUDP("udp4", nil, address)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	packetToSend := u.Message{Text: text}
	if dest != "" {
		packetToSend.Destination = &dest
	}

	// serializing the packet to send
	bytesToSend := u.ProtobufMessage(&packetToSend)

	// sending the packet over udp
	_, err = udpConn.Write(bytesToSend)
	if err != nil {
		fmt.Println(err)
	}
}

// RequestFile create and send a request for a file to a destination
func (g *Gossiper) RequestFile(name, dest, hash string) {
	// translate the string hash to byte array
	hashByte, err := hex.DecodeString(hash)
	if err != nil {
		g.Printer.Println("Error: invalid hash")
	}
	// create the data request
	req := u.DataRequest{
		Origin:      g.Name,
		Destination: dest,
		HopLimit:    u.DefaultHopLimit,
		HashValue:   hashByte,
	}

	var h u.ShaHash
	copy(h[:], hashByte)
	data := make([][]byte, 0)
	// create the new file status
	fstatus := u.FileRequestStatus{
		Name:         name,
		Destination:  dest,
		MetafileHash: h,
		MetafileOK:   false,
		ChunkCount:   0,
		Data:         data,
	}
	g.FileStatus = append(g.FileStatus, &fstatus)

	// print downloading metafile message
	g.PrintDownloadMetaFile(dest, name)

	g.RouteDataReq(req)
}

// RequestNextChunk request the next chunk to dest with hash
func (g *Gossiper) RequestNextChunk(fstatus *u.FileRequestStatus) {
	// create the new file request
	req := u.DataRequest{
		Origin:      g.Name,
		Destination: fstatus.Destination,
		HopLimit:    u.DefaultHopLimit,
		HashValue:   fstatus.PendingChunks[fstatus.ChunkCount][:],
	}
	// print downloading chunk message
	g.PrintDownloadChunk(fstatus.Destination, fstatus.Name,
		fstatus.ChunkCount+1)
	// route the packet
	g.RouteDataReq(req)
}
