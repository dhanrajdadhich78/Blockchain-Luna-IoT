package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	bc "wizeBlock/wizeNode/blockchain"
)

// TODO: refactoring
//       done: REST Server, Mutex?
//       todo: TCP Server
//       todo: blockchain, preparedTxs
//       todo: logger

type ErrorResponse struct {
	Error string `json:"error"`
}

type PreparedTransaction struct {
	From        string
	Transaction *bc.Transaction
}

type Node struct {
	restServer *RestServer

	nodeID  string
	nodeADD string
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
	return newNode
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
