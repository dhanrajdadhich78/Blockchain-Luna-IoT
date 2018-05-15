package node

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"io/ioutil"
	"net"
	"time"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
)

// DONE: divide server_tcp onto nodeserver and nodeclient
// DONE: constants and util functions to network module
// DOING: KnownNodes to nodenetwork module

const timeFormat = "15:04:05.000000"

type NodeServer struct {
	Node        *Node
	NodeAddress network.NodeAddr

	minerAddress string

	// TODO: to redesign
	blocksInTransit [][]byte
	bc              *blockchain.Blockchain
	mempool         map[string]blockchain.Transaction

	// TODO: to redesign
	conn                net.Conn
	StopMainChan        chan struct{}
	StopMainConfirmChan chan struct{}
}

func NewServer(node *Node, minerAddress string) *NodeServer {
	return &NodeServer{
		Node:            node,
		NodeAddress:     node.NodeAddress,
		minerAddress:    minerAddress,
		blocksInTransit: [][]byte{},
		mempool:         make(map[string]blockchain.Transaction),
		bc:              node.blockchain,
	}
}

// TESTS: should we support this?
//func NewServerForTest(nodeADD, nodeID, minerAddress string) *NodeServer {
//	return &NodeServer{
//		Node:            nil,
//		nodeID:          nodeID,
//		nodeADD:         nodeADD,
//		minerAddress:    minerAddress,
//		blocksInTransit: [][]byte{},
//		mempool:         make(map[string]blockchain.Transaction),
//		bc:              blockchain.NewBlockchain(nodeID),
//	}
//}

// Start a node server
func (s *NodeServer) Start(serverStartResult chan string) error {
	log.Info.Println("Prepare server to start ", s.NodeAddress.NodeAddrToString())

	ln, err := net.Listen(network.Protocol, s.Node.NodeAddress.String())
	if err != nil {
		serverStartResult <- err.Error()
		close(s.StopMainConfirmChan)

		log.Warn.Println("Fail to start port listening ", err.Error())
		return err
	}
	defer ln.Close()

	s.Node.Client.SetNodeAddress(s.NodeAddress)

	log.Info.Println("NodeServer was started")
	log.Info.Printf("nodeAddress: %s, knownNodes: %v", s.Node.NodeAddress, s.Node.Network.Nodes)
	//s.bc = s.node.blockchain

	// TODO: P2P: should we send ComVersion at the start?

	///s.Node.SendVersionToNodes([]netlib.NodeAddr{})

	log.Info.Printf("Compare node address [%s] with 0-node [%s]\n", s.Node.Client.NodeAddress, s.Node.Network.Nodes[0])
	if !s.Node.Client.NodeAddress.CompareToAddress(s.Node.Network.Nodes[0]) {
		log.Info.Printf("Send version\n")
		s.Node.Client.SendVersion(s.Node.Network.Nodes[0], s.bc.GetBestHeight())
	}

	// notify node about server started fine
	serverStartResult <- ""

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Warn.Printf("NodeServer accept error: %s", err)
			return err
		}

		// check if it is a time to stop this loop
		stop := false

		// check if a channel is still open
		// it can be closed in goroutine when receive external stop signal
		select {
		case _, ok := <-s.StopMainChan:
			if !ok {
				stop = true
			}
		default:
		}

		if stop {
			// complete all tasks, save data if needed
			ln.Close()

			close(s.StopMainConfirmChan)

			log.Info.Println("Stop Listing Network. Correct exit")
			break
		}

		go s.handleConnection(conn)
	}

	return nil
}

func (s *NodeServer) handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Warn.Printf("NodeServer read request error: %s", err)
		return
	}
	command := network.BytesToCommand(request[:network.CommandLength])
	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: Received %s command\n", s.Node.NodeID, nanonow, command)

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
		log.Warn.Println("Unknown command!")
	}

	conn.Close()
}

func (s *NodeServer) Stop() {
	if s.bc != nil && s.bc.Db != nil {
		s.bc.Db.Close()
	}
}

func (s *NodeServer) handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload network.ComAddr

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Warn.Println(err)
		return
	}

	// TODO: check this logic
	//KnownNodes = append(KnownNodes, payload.AddrList...)
	//nanonow := time.Now().Format(timeFormat)
	//fmt.Printf("nodeID: %s, %s: There are %d known nodes now!\n", s.nodeID, nanonow, len(KnownNodes))
	//for _, node := range KnownNodes {
	//	s.Node.Client.SendGetBlocks(node, s.nodeAddress)
	//}

	log.Debug.Printf("Received nodes %s", payload.AddrList)

	// TODO: check this logic
	addednodes := []network.NodeAddr{}
	for _, node := range payload.AddrList {
		if s.Node.Network.AddNodeToKnown(node) {
			addednodes = append(addednodes, node)
		}
	}

	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: There are %d known nodes now!\n", s.Node.NodeID, nanonow, len(s.Node.Network.Nodes))
	log.Debug.Printf("Send version to %d new nodes\n", len(addednodes))

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
	var payload network.ComBlock

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Warn.Println(err)
		return
	}

	blockData := payload.Block
	block := blockchain.DeserializeBlock(blockData)

	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: Received a new block!\n", s.Node.NodeID, nanonow)
	s.bc.AddBlock(block)

	log.Debug.Printf("nodeID: %s, %s: Added block %x\n", s.Node.NodeID, nanonow, block.Hash)

	// OLDTODO: add validation of block
	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		s.Node.Client.SendGetData(payload.AddrFrom, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{s.bc}
		// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
		UTXOSet.Reindex()
	}
}

func (s *NodeServer) handleInv(request []byte) {
	var buff bytes.Buffer
	var payload network.ComInv

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Warn.Println(err)
		return
	}

	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: Received inventory with %d %s\n", s.Node.NodeID, nanonow, len(payload.Items), payload.Type)
	log.Debug.Printf("len(mempool): %d\n", len(s.mempool))

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
	var payload network.ComGetBlocks

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Warn.Println(err)
		return
	}

	blocks := s.bc.GetBlockHashes()
	s.Node.Client.SendInv(payload.AddrFrom, "block", blocks)
}

func (s *NodeServer) handleGetData(request []byte) {
	var buff bytes.Buffer
	var payload network.ComGetData

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Fatal.Println(err)
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
	var payload network.ComTx

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Fatal.Println(err)
		return
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	log.Debug.Printf("handleTx: [%x] minerAddress: %s\n", tx.ID, s.minerAddress)

	// TODO: mempool should be just for miners?
	//if len(s.miningAddress) > 0 {
	log.Debug.Printf("Added to pool %d Tx: [%x]\n", len(s.mempool), tx.ID)
	s.mempool[hex.EncodeToString(tx.ID)] = tx
	//}

	//if s.nodeAddress == KnownNodes[0] {
	if s.Node.Client.NodeAddress.CompareToAddress(s.Node.Network.Nodes[0]) {
		log.Debug.Printf("nodeID: %s, knownNodes: %v\n", s.Node.NodeID, s.Node.Network.Nodes)
		for _, node := range s.Node.Network.Nodes {
			if !node.CompareToAddress(s.Node.Client.NodeAddress) &&
				!node.CompareToAddress(payload.AddFrom) {
				s.Node.Client.SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		// OLDTODO: changing count of transaction for mining
		log.Debug.Printf("minerAddress: %s, len(mempool): %d\n", s.minerAddress, len(s.mempool))
		// FIXME: len of mempool?
		if len(s.mempool) >= 1 && len(s.minerAddress) > 0 {
		MineTransactions:
			log.Debug.Println("MineTransactions...")
			var txs []*blockchain.Transaction

			for id := range s.mempool {
				tx := s.mempool[id]
				check, err := s.bc.VerifyTransaction(&tx)
				if check && err == nil {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				log.Debug.Println("All transactions are invalid! Waiting for new ones...")
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

			cbTx := blockchain.NewCoinbaseTX(s.minerAddress, "")
			txs = append(txs, cbTx)

			newBlock := s.bc.MineBlock(txs)
			UTXOSet := blockchain.UTXOSet{s.bc}
			// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
			UTXOSet.Reindex()

			nanonow := time.Now().Format(timeFormat)
			log.Debug.Printf("nodeID: %s, %s: New block is mined!", s.Node.NodeID, nanonow)
			log.Debug.Printf("New block with %d tx is mined!\n", len(txs))

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
	var payload network.ComVersion

	buff.Write(request[network.CommandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Fatal.Println(err)
		return
	}

	myBestHeight := s.bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		s.Node.Client.SendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		s.Node.Client.SendVersion(payload.AddrFrom, myBestHeight)
	}

	// sendAddr(payload.AddrFrom)
	log.Info.Printf("Check address known for %s\n", payload.AddrFrom)
	s.Node.CheckAddressKnown(payload.AddrFrom)
}
