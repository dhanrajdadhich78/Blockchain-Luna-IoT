package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	bc "wizeBlock/wizeNode/blockchain"
)

// TODO: refactoring
//       done: REST Server, Mutex?
//       todo: TCP Server
//       todo: blockchain, preparedTxs
//       todo: logger

// TODO: dataDir?
// TODO: minterAddress?
// TODO: known (other) nodes?
// TODO: NodeNet?
// TODO: NodeClient

// TODO: NodeBlockchain!
// TODO: NodeTransactions!

type PreparedTransaction struct {
	From        string
	Transaction *bc.Transaction
}

type Node struct {
	restServer *RestServer

	Network NodeNetwork
	Client  *NodeClient

	nodeADD string
	nodeID  string
	apiADD  string

	blockchain  *bc.Blockchain
	preparedTxs map[string]*PreparedTransaction

	logger *log.Logger
}

func NewNode(nodeADD, nodeID, apiADD string) *Node {
	newNode := &Node{
		nodeADD:     nodeADD,
		nodeID:      nodeID,
		apiADD:      apiADD,
		blockchain:  bc.NewBlockchain(nodeID),
		preparedTxs: make(map[string]*PreparedTransaction),
		logger: log.New(
			os.Stdout,
			"node: ",
			log.Ldate|log.Ltime,
		),
	}

	newNode.restServer = NewRestServer(newNode, apiADD)

	// HACK: KnownNodes
	newNode.Network.SetNodes([]NodeAddr{
		NodeAddr{
			Host: "localhost",
			Port: 3000,
		},
	}, true)

	newNode.InitClient()

	// HACK: Node Address
	port, _ := strconv.Atoi(nodeID)
	addr := NodeAddr{
		Host: nodeADD,
		Port: port,
	}

	newNode.Client.SetNodeAddress(addr)

	return newNode
}

func (node *Node) InitClient() error {
	if node.Client != nil {
		return nil
	}

	client := NodeClient{}
	client.Network = &node.Network
	node.Client = &client

	return nil
}

/*
 * Check if the address is known . If not then add to known
 * and send list of all addresses to that node
 */
func (node *Node) CheckAddressKnown(addr NodeAddr) {
	if !node.Network.CheckIsKnown(addr) {
		// send him all addresses
		fmt.Printf("sending list of address to %s, %s", addr.NodeAddrToString(), node.Network.Nodes)

		node.Network.AddNodeToKnown(addr)
	}
}

/*
 * Send own version to all known nodes
 */
func (node *Node) SendVersionToNodes(nodes []NodeAddr) {
	bestHeight := node.blockchain.GetBestHeight()

	if len(nodes) == 0 {
		nodes = node.Network.Nodes
	}

	for _, n := range nodes {
		if n.CompareToAddress(node.Client.NodeAddress) {
			continue
		}
		node.Client.SendVersion(n, bestHeight)
	}
}

func (node *Node) Run(minerAddress string) {
	fmt.Println("nodeADD:", node.nodeADD, "nodeID:", node.nodeID, "apiADD:", node.apiADD)

	// REST Server start
	if err := node.restServer.Start(); err != nil {
		fmt.Printf("Failed to start HTTP service: %s", err.Error())
	}

	// Node Server start
	tcpSrv := NewServer(node, minerAddress)
	go func() {
		node.log("start TCP server")
		tcpSrv.Start()
	}()

	// TODO: refactoring exits from all routines
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)
	for {
		s := <-signalCh
		if s == syscall.SIGTERM {
			node.log("stop servers")
			// FIXME
			//apiSrv.Shutdown(context.Background())
			node.restServer.Close()
			tcpSrv.Stop()
		}
	}
}

func (node *Node) log(v ...interface{}) {
	node.logger.Println(v)
}

func (node *Node) logError(err error) {
	node.log("[ERROR]", err)
}
