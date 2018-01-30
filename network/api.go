package network

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/websocket"
	m "wizeBlockchain/message"
)

func (node *Node) blocksHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(node.blockchain.Blocks)
	if err != nil {
		node.error(w, err, "failed to decode response")
		return
	}

	node.writeResponse(w, b)
}

func (node *Node) mineBlockHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Data string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		node.error(w, err, "failed to decode params")
		return
	}

	block := node.blockchain.GenerateBlock(params.Data)
	node.blockchain.AddBlock(block)
	node.log("mined block:", block.Hash)

	msg, err := node.newLatestBlockMessage()
	if err != nil {
		node.error(w, err, "failed to build message")
	}
	node.broadcast(msg)

	b, err := json.Marshal(map[string]string{
		"hash": block.Hhash(),
	})
	if err != nil {
		node.error(w, err, "failed to decode response")
		return
	}

	node.writeResponse(w, b)
}

func (node *Node) peersHandler(w http.ResponseWriter, r *http.Request) {
	peerHosts := make([]string, len(node.conns))
	for i, conn := range node.conns {
		peerHosts[i] = conn.remoteHost()
	}

	b, err := json.Marshal(peerHosts)
	if err != nil {
		node.error(w, err, "failed to decode response")
		return
	}

	node.writeResponse(w, b)
}

func (node *Node) addPeerHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Peer string `json:"peer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		node.error(w, err, "failed to decode params")
		return
	}

	ws, err := websocket.Dial(params.Peer, "", *p2pOrigin)
	if err != nil {
		node.error(w, err, "failed to connect to peer")
		return
	}

	conn := NewConn(ws)
	node.log("connect to peer:", conn.remoteHost())
	node.addConn(conn)
	go node.p2pHandler(conn)

	if err := node.send(conn, m.NewQueryLatestMessage()); err != nil {
		node.logError(err)
	}

	node.peersHandler(w, r)
}
