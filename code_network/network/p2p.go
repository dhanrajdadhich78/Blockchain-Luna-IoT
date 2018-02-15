package network

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"golang.org/x/net/websocket"
	b "wizeBlock/code_network/block"
	bc "wizeBlock/code_network/blockchain"
	m "wizeBlock/code_network/message"
)

func (node *Node) addConn(conn *Conn) {
	node.mu.Lock()
	defer node.mu.Unlock()

	node.conns = append(node.conns, conn)
}

func (node *Node) deleteConn(id int64) {
	node.mu.Lock()
	defer node.mu.Unlock()

	conns := []*Conn{}
	for _, conn := range node.conns {
		if conn.id != id {
			conns = append(conns, conn)
		}
	}

	node.conns = conns
}

func (node *Node) disconnectPeer(conn *Conn) {
	defer conn.Close()
	node.log("disconnect peer:", conn.remoteHost())
	node.deleteConn(conn.id)
}

func (node *Node) newLatestBlockMessage() (*m.Message, error) {
	return m.NewBlocksMessage(b.Blocks{node.blockchain.GetLatestBlock()})
}

func (node *Node) newAllBlocksMessage() (*m.Message, error) {
	return m.NewBlocksMessage(node.blockchain.Blocks)
}

func (node *Node) broadcast(msg *m.Message) {
	for _, conn := range node.conns {
		if err := node.send(conn, msg); err != nil {
			node.logError(err)
		}
	}
}

func (node *Node) send(conn *Conn, msg *m.Message) error {
	node.log(fmt.Sprintf(
		"send %s message to %s",
		msg.Type.Name(), conn.remoteHost(),
	))

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return websocket.Message.Send(conn.Conn, b)
}

func (node *Node) handleBlocksResponse(conn *Conn, msg *m.Message) error {
	var blocks b.Blocks
	if err := json.Unmarshal([]byte(msg.Data), &blocks); err != nil {
		return err
	}
	sort.Sort(blocks)

	latestBlock := blocks[len(blocks)-1]
	if latestBlock.Index > node.blockchain.GetLatestBlock().Index {
		if node.blockchain.GetLatestBlock().Hash == latestBlock.PreviousHash {
			if b.IsValidBlock(latestBlock, node.blockchain.GetLatestBlock()) {
				node.blockchain.AddBlock(latestBlock)

				msg, err := node.newLatestBlockMessage()
				if err != nil {
					return err
				}

				node.broadcast(msg)
			}
		} else if len(blocks) == 1 {
			node.broadcast(m.NewQueryAllMessage())
		} else {
			bc := bc.NewBlockchain()
			bc.ReplaceBlocks(blocks)
			if bc.IsValid() {
				node.blockchain.ReplaceBlocks(blocks)

				msg, err := node.newLatestBlockMessage()
				if err != nil {
					return err
				}

				node.broadcast(msg)
			}
		}
	}

	return nil
}

func (node *Node) p2pHandler(conn *Conn) {
	for {
		var b []byte
		if err := websocket.Message.Receive(conn.Conn, &b); err != nil {
			if err == io.EOF {
				node.disconnectPeer(conn)
				break
			}
			node.logError(err)
			continue
		}

		var msg m.Message
		if err := json.Unmarshal(b, &msg); err != nil {
			node.logError(err)
			continue
		}

		node.log(fmt.Sprintf(
			"received %s message from %s",
			msg.Type.Name(), conn.remoteHost(),
		))

		switch msg.Type {
		case m.MessageTypeQueryLatest:
			msg, err := node.newLatestBlockMessage()
			if err != nil {
				node.logError(err)
				continue
			}
			if err := node.send(conn, msg); err != nil {
				node.logError(err)
			}
		case m.MessageTypeQueryAll:
			msg, err := node.newAllBlocksMessage()
			if err != nil {
				node.logError(err)
				continue
			}
			if err := node.send(conn, msg); err != nil {
				node.logError(err)
			}
		case m.MessageTypeBlocks:
			if err := node.handleBlocksResponse(conn, &msg); err != nil {
				node.logError(err)
			}
		default:
			node.logError(m.ErrUnknownMessageType)
		}
	}
}
