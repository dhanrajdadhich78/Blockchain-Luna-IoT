package node

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
)

type NodeServerRequest struct {
	Node        *Node
	Server      *NodeServer
	Request     []byte
	RequestIP   string
	HasResponse bool
	Response    []byte
}

func (self *NodeServerRequest) Init() {
	self.HasResponse = false
	self.Response = nil
}

// Reads and parses request from network data
func (self *NodeServerRequest) parseRequestData(payload interface{}) error {
	var buff bytes.Buffer

	buff.Write(self.Request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(payload)

	if err != nil {
		return fmt.Errorf("Parse request: " + err.Error())
	}

	return nil
}

func (self *NodeServerRequest) handleAddr() error {
	var payload network.ComAddr
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	// TODO: check this logic
	//KnownNodes = append(KnownNodes, payload.AddrList...)
	//nanonow := time.Now().Format(timeFormat)
	//fmt.Printf("nodeID: %s, %s: There are %d known nodes now!\n", s.nodeID, nanonow, len(KnownNodes))
	//for _, node := range KnownNodes {
	//	s.Node.Client.SendGetBlocks(node, s.nodeAddress)
	//}

	log.Info.Printf("Received nodes %s", payload.AddrList)

	// TODO: check this logic
	addednodes := []network.NodeAddr{}
	for _, node := range payload.AddrList {
		if self.Node.Network.AddNodeToKnown(node) {
			addednodes = append(addednodes, node)
		}
	}

	nanonow := time.Now().Format(timeFormat)
	log.Info.Printf("nodeID: %s, %s: There are %d known nodes now!\n", self.Node.NodeID, nanonow, len(self.Node.Network.Nodes))
	log.Info.Printf("Send version to %d new nodes\n", len(addednodes))

	if len(addednodes) > 0 {
		// send own version to all new found nodes. maybe they have some more blocks
		// and they will add me to known nodes after this
		self.Node.SendVersionToNodes(addednodes)
	}

	return nil
}

// OLDTODO: проверять остаток на балансе с учетом незамайненых транзакций,
//          во избежание двойного использования выходов
func (self *NodeServerRequest) handleBlock() error {
	var payload network.ComBlock
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	blockData := payload.Block
	block := blockchain.DeserializeBlock(blockData)

	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: Received a new block!\n", self.Node.NodeID, nanonow)
	self.Server.bc.AddBlock(block)

	log.Debug.Printf("nodeID: %s, %s: Added block %x\n", self.Node.NodeID, nanonow, block.Hash)

	// OLDTODO: add validation of block
	if len(self.Server.blocksInTransit) > 0 {
		blockHash := self.Server.blocksInTransit[0]
		self.Node.Client.SendGetData(payload.AddrFrom, "block", blockHash)

		self.Server.blocksInTransit = self.Server.blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{self.Server.bc}
		// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
		UTXOSet.Reindex()
	}

	return nil
}

func (self *NodeServerRequest) handleInv() error {
	var payload network.ComInv
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: Received inventory with %d %s\n", self.Node.NodeID, nanonow, len(payload.Items), payload.Type)
	log.Debug.Printf("len(mempool): %d\n", len(self.Server.mempool))

	if payload.Type == "block" {
		self.Server.blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		self.Node.Client.SendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range self.Server.blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		self.Server.blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if self.Server.mempool[hex.EncodeToString(txID)].ID == nil {
			self.Node.Client.SendGetData(payload.AddrFrom, "tx", txID)
		}
	}

	return nil
}

func (self *NodeServerRequest) handleGetBlocks() error {
	var payload network.ComGetBlocks
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	blocks := self.Server.bc.GetBlockHashes()
	self.Node.Client.SendInv(payload.AddrFrom, "block", blocks)

	return nil
}

func (self *NodeServerRequest) handleGetData() error {
	var payload network.ComGetData
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	if payload.Type == "block" {
		block, err := self.Server.bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return err
		}

		self.Node.Client.SendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := self.Server.mempool[txID]

		self.Node.Client.SendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}

	return nil
}

func (self *NodeServerRequest) handleTx() error {
	var payload network.ComTx
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	log.Debug.Printf("handleTx: [%x] minerAddress: %s\n", tx.ID, self.Server.minerAddress)

	// TODO: mempool should be just for miners?
	//if len(s.miningAddress) > 0 {
	log.Debug.Printf("Added to pool %d Tx: [%x]\n", len(self.Server.mempool), tx.ID)
	self.Server.mempool[hex.EncodeToString(tx.ID)] = tx
	//}

	//if s.nodeAddress == KnownNodes[0] {
	if self.Node.Client.NodeAddress.CompareToAddress(self.Node.Network.Nodes[0]) {
		log.Debug.Printf("nodeID: %s, knownNodes: %v\n", self.Node.NodeID, self.Node.Network.Nodes)
		for _, node := range self.Node.Network.Nodes {
			if !node.CompareToAddress(self.Node.Client.NodeAddress) &&
				!node.CompareToAddress(payload.AddFrom) {
				self.Node.Client.SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		// OLDTODO: changing count of transaction for mining
		log.Debug.Printf("minerAddress: %s, len(mempool): %d\n", self.Server.minerAddress, len(self.Server.mempool))
		// FIXME: len of mempool?
		if len(self.Server.mempool) >= 1 && len(self.Server.minerAddress) > 0 {
		MineTransactions:
			log.Debug.Println("MineTransactions...")
			var txs []*blockchain.Transaction

			for id := range self.Server.mempool {
				tx := self.Server.mempool[id]
				check, err := self.Server.bc.VerifyTransaction(&tx)
				if check && err == nil {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				log.Debug.Println("All transactions are invalid! Waiting for new ones...")
				return nil
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

			cbTx := blockchain.NewCoinbaseTX(self.Server.minerAddress, "")
			txs = append(txs, cbTx)

			newBlock := self.Server.bc.MineBlock(txs)
			UTXOSet := blockchain.UTXOSet{self.Server.bc}
			// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
			UTXOSet.Reindex()

			nanonow := time.Now().Format(timeFormat)
			log.Debug.Printf("nodeID: %s, %s: New block is mined!", self.Node.NodeID, nanonow)
			log.Debug.Printf("New block with %d tx is mined!\n", len(txs))

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(self.Server.mempool, txID)
			}

			for _, node := range self.Node.Network.Nodes {
				if !node.CompareToAddress(self.Node.Client.NodeAddress) {
					self.Node.Client.SendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(self.Server.mempool) > 0 {
				goto MineTransactions
			}
		}
	}

	return nil
}

func (self *NodeServerRequest) handleVersion() error {
	var payload network.ComVersion
	err := self.parseRequestData(&payload)
	if err != nil {
		return err
	}

	myBestHeight := self.Server.bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		log.Info.Printf("Request blocks from %s\n", payload.AddrFrom)

		self.Node.Client.SendGetBlocks(payload.AddrFrom)

	} else if myBestHeight > foreignerBestHeight {
		log.Info.Printf("Send my version back to %s\n", payload.AddrFrom)

		self.Node.Client.SendVersion(payload.AddrFrom, myBestHeight)

	} else {
		log.Info.Printf("Their blockchain is same as my for %s\n", payload.AddrFrom)
	}

	log.Info.Printf("Check address known for %s\n", payload.AddrFrom)
	self.Server.Node.CheckAddressKnown(payload.AddrFrom)

	return nil
}
