package node

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
)

// TODO: rethink with NewNodeServer and Start/Stop
// TODO: rethink with handleConnection and readRequest
// TODO: rethink with CloneNode and multiple goroutines

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

func NewNodeServer(node *Node, minerAddress string) *NodeServer {
	return &NodeServer{
		Node:                node,
		NodeAddress:         node.NodeAddress,
		minerAddress:        minerAddress,
		blocksInTransit:     [][]byte{},
		mempool:             make(map[string]blockchain.Transaction),
		bc:                  node.blockchain,
		StopMainChan:        make(chan struct{}),
		StopMainConfirmChan: make(chan struct{}),
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
	log.Info.Println("Prepare Node Server to start ", s.NodeAddress)

	ln, err := net.Listen(network.Protocol, s.Node.NodeAddress.String())
	if err != nil {
		serverStartResult <- err.Error()
		//close(s.StopMainConfirmChan)

		log.Warn.Println("Fail to start port listening ", err.Error())
		return err
	}
	defer ln.Close()

	s.Node.Client.SetNodeAddress(s.NodeAddress)

	log.Info.Printf("Node Server [%s] was started, knownNodes: %v",
		s.Node.NodeAddress, s.Node.Network.Nodes)
	//s.bc = s.node.blockchain

	s.Node.SendVersionToNodes([]network.NodeAddr{})

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

			//close(s.StopMainConfirmChan)

			log.Info.Println("Stop Listing Network. Correct exit")
			break
		}

		go s.handleConnection(conn)
	}

	return nil
}

func (s *NodeServer) Stop() {
	if s.bc != nil && s.bc.Db != nil {
		s.bc.Db.Close()
	}
}

func (s *NodeServer) readRequest(conn net.Conn) (command string, databuffer []byte, err error) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Warn.Printf("NodeServer read request error: %s", err)
		return
	}

	command = network.BytesToCommand(request[:network.CommandLength])
	databuffer = request[network.CommandLength:]
	return
}

func (s *NodeServer) handleConnection(conn net.Conn) {
	starttime := time.Now().UnixNano()
	log.Debug.Println("New command. Start reading")

	command, request, err := s.readRequest(conn)
	if err != nil {
		s.sendErrorBack(conn, fmt.Errorf("Network Data Reading Error: "+err.Error()))
		conn.Close()
		return
	}

	log.Debug.Printf("Received %s command", command)

	nanonow := time.Now().Format(timeFormat)
	log.Debug.Printf("nodeID: %s, %s: Received %s command\n", s.Node.NodeID, nanonow, command)

	// FIXME: with constructor
	requestObj := NodeServerRequest{}
	// HACK: should we clone node?
	requestObj.Node = s.CloneNode()
	requestObj.Request = request[:]
	requestObj.Server = s
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		requestObj.RequestIP = addr.IP.String()
	}
	request = nil

	log.Debug.Printf("RequestObj: %+v\n", requestObj)

	var rerr error
	switch command {
	case "addr":
		rerr = requestObj.handleAddr()
	case "block":
		rerr = requestObj.handleBlock()
	case "inv":
		rerr = requestObj.handleInv()
	case "getblocks":
		rerr = requestObj.handleGetBlocks()
	case "getdata":
		rerr = requestObj.handleGetData()
	case "tx":
		rerr = requestObj.handleTx()
	case "version":
		rerr = requestObj.handleVersion()
	default:
		rerr = fmt.Errorf("Unknown command!")
	}

	if rerr != nil {
		log.Info.Println("Network Command Handle Error: ", rerr.Error())
		if requestObj.HasResponse {
			// return error to the client
			// first byte is bool false to indicate there was error
			s.sendErrorBack(conn, rerr)
		}
	}

	// TODO: add sending response
	//	if requestObj.HasResponse && requestObj.Response != nil && rerr == nil {
	//		// send this response back
	//		// first byte is bool true to indicate request was success
	//		dataresponse := append([]byte{1}, requestObj.Response...)
	//		log.Info.Printf("Responding %d bytes\n", len(dataresponse))
	//		_, err := conn.Write(dataresponse)
	//		if err != nil {
	//			log.Warn.Println("Sending response error: ", err.Error())
	//		}
	//	}

	duration := time.Since(time.Unix(0, starttime))
	ms := duration.Nanoseconds() / int64(time.Millisecond)
	log.Debug.Printf("Complete processing %s command. Time: %d ms\n", command, ms)

	conn.Close()
}

func (s *NodeServer) sendErrorBack(conn net.Conn, err error) {
	log.Info.Println("Sending back error message: ", err.Error())

	payload, err := network.GobEncode(err.Error())
	if err == nil {
		dataresponse := append([]byte{0}, payload...)
		log.Info.Printf("Responding %d bytes as error message\n", len(dataresponse))
		_, err = conn.Write(dataresponse)
		if err != nil {
			log.Warn.Println("Sending response error: ", err.Error())
		}
	}
}

/*
* Creates clone of a node object. We use this in case if we need separate object
* for a routine. This prevents conflicts of pointers in different routines
 */
func (s *NodeServer) CloneNode() *Node {
	originnode := s.Node

	// FIXME: should we just clone not create new object?
	//node := NewNode(originnode.NodeID, originnode.NodeAddress, originnode.apiAddr, s.minerAddress)
	node := Node{
		NodeID:      originnode.NodeID,
		NodeAddress: originnode.NodeAddress,
		blockchain:  originnode.blockchain,
	}

	node.Init()
	node.Client.SetNodeAddress(s.NodeAddress)
	// set list of nodes and skip loading default if this is empty list
	node.InitNetwork(originnode.Network.Nodes, true)

	return &node
}
