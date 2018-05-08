package app

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	bc "wizeBlock/wizeNode/blockchain"
)

// TODO: refactoring
//       done: REST Server, Mutex?
//       doing: TCP Server
//       todo: blockchain, preparedTxs
//       doing: logger

// TODO: dataDir?
// TODO: minterAddress?
// DOING: known (other) nodes - NodeNetwork
// DOING: Network?
// DOING: NodeClient

// TODO: NodeBlockchain!
// TODO: NodeTransactions!

type PreparedTransaction struct {
	From        string
	Transaction *bc.Transaction
}

type Node struct {
	Network NodeNetwork
	Client  *NodeClient

	Rest *RestServer

	// FIXME: to delete
	nodeADD string
	nodeID  string
	apiADD  string

	// FIXME: NodeBlockchain, NodeTransactions
	blockchain  *bc.Blockchain
	preparedTxs map[string]*PreparedTransaction
}

func NewNode(nodeADD, nodeID, apiADD string) *Node {
	newNode := &Node{
		nodeADD:     nodeADD,
		nodeID:      nodeID,
		apiADD:      apiADD,
		blockchain:  bc.NewBlockchain(nodeID),
		preparedTxs: make(map[string]*PreparedTransaction),
	}

	newNode.Rest = NewRestServer(newNode, apiADD)

	// HACK: KnownNodes
	newNode.Network.SetNodes([]NodeAddr{
		NodeAddr{
			Host: "wize1",
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
		LogDebug.Printf("sending list of address to %s, %s", addr.NodeAddrToString(), node.Network.Nodes)

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
	LogDebug.Printf("nodeADD: %s, nodeID: %s apiADD: %s", node.nodeADD, node.nodeID, node.apiADD)

	// REST Server start
	if err := node.Rest.Start(); err != nil {
		LogFatal.Printf("Failed to start HTTP service: %s", err)
	}

	// Node Server start
	tcpSrv := NewServer(node, minerAddress)
	go func() {
		LogInfo.Println("Start NodeServer")
		tcpSrv.Start()
	}()

	// TODO: refactoring exits from all routines
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)
	for {
		s := <-signalCh
		if s == syscall.SIGTERM {
			LogInfo.Println("Stop servers")
			// FIXME
			//apiSrv.Shutdown(context.Background())
			node.Rest.Close()
			tcpSrv.Stop()
		}
	}
}
