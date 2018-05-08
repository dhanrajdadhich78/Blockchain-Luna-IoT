package app

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	b "wizeBlock/wizeNode/blockchain"
)

// DONE: divide server_tcp onto nodeserver and nodeclient
// DONE: constants and util functions to network module
// DOING: KnownNodes to nodenetwork module

const timeFormat = "15:04:05.000000"

type NodeServer struct {
	Node *Node

	// TODO: to delete
	nodeID        string
	nodeADD       string
	nodeAddress   string
	miningAddress string

	// TODO: to redesign
	blocksInTransit [][]byte
	mempool         map[string]b.Transaction

	// TODO: to redesign
	ln   net.Listener
	conn net.Conn
	bc   *b.Blockchain
}

func NewServer(node *Node, minerAddress string) *NodeServer {
	return &NodeServer{
		Node:            node,
		nodeID:          node.nodeID,
		nodeADD:         node.nodeADD,
		nodeAddress:     fmt.Sprintf("%s:%s", node.nodeADD, node.nodeID),
		miningAddress:   minerAddress,
		blocksInTransit: [][]byte{},
		mempool:         make(map[string]b.Transaction),
		bc:              node.blockchain,
	}
}

// TODO: should we support this?
func NewServerForTest(nodeADD, nodeID, minerAddress string) *NodeServer {
	return &NodeServer{
		Node:            nil,
		nodeID:          nodeID,
		nodeADD:         nodeADD,
		nodeAddress:     fmt.Sprintf("localhost:%s", nodeID),
		miningAddress:   minerAddress,
		blocksInTransit: [][]byte{},
		mempool:         make(map[string]b.Transaction),
		bc:              b.NewBlockchain(nodeID),
	}
}

// StartServer starts a node
func (s *NodeServer) Start() {
	var err error
	s.ln, err = net.Listen(Protocol, s.nodeAddress)
	if err != nil {
		LogWarn.Printf("NodeServer start error: %s", err)
		return
	}
	// TODO: should we uncomment or delete this?
	//defer server.ln.Close()

	LogInfo.Println("NodeServer was started")
	LogDebug.Printf("nodeAddress: %s, knownNodes: %v", s.nodeAddress, s.Node.Network.Nodes)
	//s.bc = s.node.blockchain

	// TODO: should we send ComVersion at the start?
	//if s.nodeAddress != KnownNodes[0] {
	//	s.Node.Client.SendVersion(KnownNodes[0], s.nodeAddress, s.bc)
	//}

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			LogWarn.Printf("NodeServer accept error: %s", err)
			break
		}
		go s.handleConnection(conn)
	}
}

func (s *NodeServer) handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		LogWarn.Printf("NodeServer read request error: %s", err)
		return
	}
	command := BytesToCommand(request[:CommandLength])
	nanonow := time.Now().Format(timeFormat)
	LogDebug.Printf("nodeID: %s, %s: Received %s command\n", s.nodeID, nanonow, command)

	switch command {
	case "addr":
		s.handleAddr(request)
	case "block":
		s.handleBlock(request)
	case "inv":
		s.handleInv(request)
	case "getblocks":
		s.handleGetBlocks(request)
	case "getdata":
		s.handleGetData(request)
	case "tx":
		s.handleTx(request)
	case "version":
		s.handleVersion(request)
	default:
		LogWarn.Println("Unknown command!")
	}

	conn.Close()
}

func (s *NodeServer) Stop() {
	if s.ln != nil {
		s.ln.Close()
	}
	if s.bc != nil && s.bc.Db != nil {
		s.bc.Db.Close()
	}
}

func (s *NodeServer) handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload ComAddr

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogWarn.Println(err)
		return
	}

	// TODO: check this logic
	//KnownNodes = append(KnownNodes, payload.AddrList...)
	//nanonow := time.Now().Format(timeFormat)
	//fmt.Printf("nodeID: %s, %s: There are %d known nodes now!\n", s.nodeID, nanonow, len(KnownNodes))
	//for _, node := range KnownNodes {
	//	s.Node.Client.SendGetBlocks(node, s.nodeAddress)
	//}

	LogDebug.Printf("Received nodes %s", payload.AddrList)

	// TODO: check this logic
	addednodes := []NodeAddr{}
	for _, node := range payload.AddrList {
		if s.Node.Network.AddNodeToKnown(node) {
			addednodes = append(addednodes, node)
		}
	}

	nanonow := time.Now().Format(timeFormat)
	LogDebug.Printf("nodeID: %s, %s: There are %d known nodes now!\n", s.nodeID, nanonow, len(s.Node.Network.Nodes))
	LogDebug.Printf("Send version to %d new nodes\n", len(addednodes))

	if len(addednodes) > 0 {
		// send own version to all new found nodes. maybe they have some more blocks
		// and they will add me to known nodes after this
		s.Node.SendVersionToNodes(addednodes)
	}
}

// OLDTODO: проверять остаток на балансе с учетом незамайненых транзакций,
//          во избежание двойного использования выходов
func (s *NodeServer) handleBlock(request []byte) {
	var buff bytes.Buffer
	var payload ComBlock

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogWarn.Println(err)
		return
	}

	blockData := payload.Block
	block := b.DeserializeBlock(blockData)

	nanonow := time.Now().Format(timeFormat)
	LogDebug.Printf("nodeID: %s, %s: Received a new block!\n", s.nodeID, nanonow)
	s.bc.AddBlock(block)

	LogDebug.Printf("nodeID: %s, %s: Added block %x\n", s.nodeID, nanonow, block.Hash)

	// OLDTODO: add validation of block
	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		s.Node.Client.SendGetData(payload.AddrFrom, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := b.UTXOSet{s.bc}
		// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
		UTXOSet.Reindex()
	}
}

func (s *NodeServer) handleInv(request []byte) {
	var buff bytes.Buffer
	var payload ComInv

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogWarn.Println(err)
		return
	}

	nanonow := time.Now().Format(timeFormat)
	LogDebug.Printf("nodeID: %s, %s: Received inventory with %d %s\n", s.nodeID, nanonow, len(payload.Items), payload.Type)
	LogDebug.Printf("len(mempool): %d\n", len(s.mempool))

	if payload.Type == "block" {
		s.blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		s.Node.Client.SendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range s.blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		s.blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if s.mempool[hex.EncodeToString(txID)].ID == nil {
			s.Node.Client.SendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func (s *NodeServer) handleGetBlocks(request []byte) {
	var buff bytes.Buffer
	var payload ComGetBlocks

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogWarn.Println(err)
		return
	}

	blocks := s.bc.GetBlockHashes()
	s.Node.Client.SendInv(payload.AddrFrom, "block", blocks)
}

func (s *NodeServer) handleGetData(request []byte) {
	var buff bytes.Buffer
	var payload ComGetData

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogFatal.Println(err)
		return
	}

	if payload.Type == "block" {
		block, err := s.bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		s.Node.Client.SendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := s.mempool[txID]

		s.Node.Client.SendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}
}

func (s *NodeServer) handleTx(request []byte) {
	var buff bytes.Buffer
	var payload ComTx

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogFatal.Println(err)
		return
	}

	txData := payload.Transaction
	tx := b.DeserializeTransaction(txData)
	LogDebug.Printf("handleTx: [%x] miningAddress: %s\n", tx.ID, s.miningAddress)

	// TODO: mempool should be just for miners?
	//if len(s.miningAddress) > 0 {
	LogDebug.Printf("Added to pool %d Tx: [%x]\n", len(s.mempool), tx.ID)
	s.mempool[hex.EncodeToString(tx.ID)] = tx
	//}

	//if s.nodeAddress == KnownNodes[0] {
	if s.Node.Client.NodeAddress.CompareToAddress(s.Node.Network.Nodes[0]) {
		LogDebug.Printf("nodeID: %s, knownNodes: %v\n", s.nodeID, s.Node.Network.Nodes)
		for _, node := range s.Node.Network.Nodes {
			if !node.CompareToAddress(s.Node.Client.NodeAddress) &&
				!node.CompareToAddress(payload.AddFrom) {
				s.Node.Client.SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		// OLDTODO: changing count of transaction for mining
		LogDebug.Printf("miningAddress: %s, len(mempool): %d\n", s.miningAddress, len(s.mempool))
		// FIXME: len of mempool?
		if len(s.mempool) >= 1 && len(s.miningAddress) > 0 {
		MineTransactions:
			LogDebug.Println("MineTransactions...")
			var txs []*b.Transaction

			for id := range s.mempool {
				tx := s.mempool[id]
				check, err := s.bc.VerifyTransaction(&tx)
				if check && err == nil {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				LogDebug.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			// TODO: fix transactions with minenow=false, count = 2? more?
			// TODO: we should check all transactions from the first to the last in time
			// TODO: check all transactions by timestamp
			/*
				var inputTxID, check = "", ""
				for index, tx := range txs {
					fmt.Printf("Index: %d, Timestamp: %d, TxID: %x\n", index, tx.Timestamp, tx.ID)
					if len(tx.Vin) > 0 {
						for vi, vin := range tx.Vin {
							fmt.Printf("  Input  %d, TxID: %x, Out: %d\n", vi, vin.Txid, vin.Vout)
						}
					}
					if len(tx.Vout) > 0 {
						for vo, vout := range tx.Vout {
							fmt.Printf("  Output %d, PubKeyHash: %x, Out: %d\n", vo, vout.PubKeyHash, vout.Value)
						}
					}

					if len(tx.Vin) > 0 && inputTxID == "" {
						inputTxID = hex.EncodeToString(tx.Vin[0].Txid)
						fmt.Printf("    inputTxID: %s\n", inputTxID)
					} else if len(tx.Vin) > 0 {
						check = hex.EncodeToString(tx.Vin[0].Txid)
						fmt.Printf("    check:     %s\n", check)
						if check == inputTxID {
							fmt.Printf("    Equals!\n")

							// FIXME
							fmt.Printf("    Fix: %x to %x\n", tx.Vin[0].Txid, txs[index-1].ID)
							//tx.Vin[0].Txid = txs[index-1].ID

							inputTxID = check
						}
					}
				}
			*/

			cbTx := b.NewCoinbaseTX(s.miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := s.bc.MineBlock(txs)
			UTXOSet := b.UTXOSet{s.bc}
			// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
			UTXOSet.Reindex()

			nanonow := time.Now().Format(timeFormat)
			LogDebug.Printf("nodeID: %s, %s: New block is mined!", s.nodeID, nanonow)
			LogDebug.Printf("New block with %d tx is mined!\n", len(txs))

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(s.mempool, txID)
			}

			for _, node := range s.Node.Network.Nodes {
				if !node.CompareToAddress(s.Node.Client.NodeAddress) {
					s.Node.Client.SendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(s.mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func (s *NodeServer) handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload ComVersion

	buff.Write(request[CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		LogFatal.Println(err)
		return
	}

	myBestHeight := s.bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		s.Node.Client.SendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		s.Node.Client.SendVersion(payload.AddrFrom, myBestHeight)
	}

	s.Node.CheckAddressKnown(payload.AddrFrom)
}
