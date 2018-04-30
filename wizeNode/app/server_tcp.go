package app

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"os"
	b "wizeBlock/wizeNode/blockchain"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
const timeFormat = "03:04:05.000000"

var KnownNodes = []string{os.Getenv("MASTERNODE") + ":3000"} //TODO: change to valid nodes in production

type TCPServer struct {
	node          *Node
	nodeID        string
	nodeADD       string
	nodeAddress   string
	miningAddress string

	blocksInTransit [][]byte
	mempool         map[string]b.Transaction

	ln   net.Listener
	conn net.Conn
	bc   *b.Blockchain
}

func NewServer(node *Node, minerAddress string) *TCPServer {
	return &TCPServer{
		node:            node,
		nodeID:          node.nodeID,
		nodeADD:         node.nodeADD,
		nodeAddress:     fmt.Sprintf("%s:%s", node.nodeADD, node.nodeID),
		miningAddress:   minerAddress,
		blocksInTransit: [][]byte{},
		mempool:         make(map[string]b.Transaction),
		bc:              node.blockchain,
	}
}

func NewServerForTest(nodeADD, nodeID, minerAddress string) *TCPServer {
	return &TCPServer{
		node:            nil,
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
func (s *TCPServer) Start() {
	fmt.Println("TCPServer Start")

	var err error
	s.ln, err = net.Listen(protocol, s.nodeAddress)
	if err != nil {
		fmt.Println(err)
		//log.Panic(err)
	}
	//defer server.ln.Close()

	fmt.Println("A nodeAddress:", s.nodeAddress, "knownNodes:", KnownNodes)
	//s.bc = s.node.blockchain

	if s.nodeAddress != KnownNodes[0] {
		sendVersion(KnownNodes[0], s.nodeAddress, s.bc)
	}

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	nanonow := time.Now().Format(timeFormat)
	fmt.Printf("nodeID: %s, %s: Received %s command\n", s.nodeID, nanonow, command)

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
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

func (s *TCPServer) Stop() {
	if s.ln != nil {
		s.ln.Close()
	}
	if s.bc != nil && s.bc.Db != nil {
		s.bc.Db.Close()
	}
}

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func nodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func sendData(address string, data []byte) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		fmt.Printf("%s is not available\n", address)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != address {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func requestBlocks(nodeAddress string) {
	for _, node := range KnownNodes {
		sendGetBlocks(node, nodeAddress)
	}
}

func sendAddr(address, nodeAddress string) {
	nodes := addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(address, request)
}

func sendBlock(address, nodeAddress string, b *b.Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(address, request)
}

func sendInv(address, nodeAddress, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func sendGetBlocks(address, nodeAddress string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

func sendGetData(address, nodeAddress, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func SendTx(address, nodeAddress string, tnx *b.Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(address, request)
}

func sendVersion(address, nodeAddress string, bc *b.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzzion{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(address, request)
}

func (s *TCPServer) handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	nanonow := time.Now().Format(timeFormat)
	fmt.Printf("nodeID: %s, %s: There are %d known nodes now!\n", s.nodeID, nanonow, len(KnownNodes))
	requestBlocks(s.nodeAddress)
}

// OLDTODO: проверять остаток на балансе с учетом незамайненых транзакций,
//          во избежание двойного использования выходов
func (s *TCPServer) handleBlock(request []byte) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := b.DeserializeBlock(blockData)

	nanonow := time.Now().Format(timeFormat)
	fmt.Printf("nodeID: %s, %s: Received a new block!\n", s.nodeID, nanonow)
	s.bc.AddBlock(block)

	fmt.Printf("nodeID: %s, %s: Added block %x\n", s.nodeID, nanonow, block.Hash)

	// OLDTODO: add validation of block
	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		sendGetData(payload.AddrFrom, s.nodeAddress, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := b.UTXOSet{s.bc}
		// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
		UTXOSet.Reindex()
	}
}

func (s *TCPServer) handleInv(request []byte) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	nanonow := time.Now().Format(timeFormat)
	fmt.Printf("nodeID: %s, %s: Received inventory with %d %s\n", s.nodeID, nanonow, len(payload.Items), payload.Type)
	//fmt.Printf("len(mempool): %d\n", len(s.mempool))

	if payload.Type == "block" {
		s.blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, s.nodeAddress, "block", blockHash)

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
			sendGetData(payload.AddrFrom, s.nodeAddress, "tx", txID)
		}
	}
}

func (s *TCPServer) handleGetBlocks(request []byte) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := s.bc.GetBlockHashes()
	sendInv(payload.AddrFrom, s.nodeAddress, "block", blocks)
}

func (s *TCPServer) handleGetData(request []byte) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := s.bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, s.nodeAddress, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := s.mempool[txID]

		SendTx(payload.AddrFrom, s.nodeAddress, &tx)
		// delete(mempool, txID)
	}
}

func (s *TCPServer) handleTx(request []byte) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := b.DeserializeTransaction(txData)
	s.mempool[hex.EncodeToString(tx.ID)] = tx

	if s.nodeAddress == KnownNodes[0] {
		fmt.Printf("nodeID: %s, knownNodes: %v\n", s.nodeID, KnownNodes)
		for _, node := range KnownNodes {
			if node != s.nodeAddress && node != payload.AddFrom {
				sendInv(node, s.nodeAddress, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		// OLDTODO: changing count of transaction for mining
		fmt.Printf("miningAddress: %s, len(mempool): %d\n", s.miningAddress, len(s.mempool))
		if len(s.mempool) >= 2 && len(s.miningAddress) > 0 {
		MineTransactions:
			fmt.Println("MineTransactions...")
			var txs []*b.Transaction

			for id := range s.mempool {
				tx := s.mempool[id]
				check, err := s.bc.VerifyTransaction(&tx)
				if check && err == nil {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := b.NewCoinbaseTX(s.miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := s.bc.MineBlock(txs)
			UTXOSet := b.UTXOSet{s.bc}
			// OLDTODO: use UTXOSet.Update() instead UTXOSet.Reindex
			UTXOSet.Reindex()

			nanonow := time.Now().Format(timeFormat)
			fmt.Printf("nodeID: %s, %s: New block is mined!", s.nodeID, nanonow)
			fmt.Printf("New block with %d tx is mined!\n", len(txs))

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(s.mempool, txID)
			}

			for _, node := range KnownNodes {
				if node != s.nodeAddress {
					sendInv(node, s.nodeAddress, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(s.mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func (s *TCPServer) handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload verzzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := s.bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom, s.nodeAddress)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, s.nodeAddress, s.bc)
	}

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		fmt.Printf("A new node %s is connected\n", payload.AddrFrom)
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}
}
