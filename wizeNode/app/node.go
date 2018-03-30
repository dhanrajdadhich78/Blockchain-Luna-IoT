package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/betacraft/yaag/middleware"
	"github.com/betacraft/yaag/yaag"
	"github.com/gorilla/mux"

	bc "wizeBlock/wizeNode/blockchain"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Node struct {
	*http.ServeMux

	blockchain  *bc.Blockchain
	preparedTxs map[string]*bc.Transaction

	mu      sync.RWMutex
	logger  *log.Logger
	apiAddr string
	nodeID  string
	nodeADD string
}

func NewNode(nodeID string) *Node {
	return &Node{
		blockchain:  bc.NewBlockchain(nodeID),
		preparedTxs: make(map[string]*bc.Transaction),
		mu:          sync.RWMutex{},
		logger: log.New(
			os.Stdout,
			"node: ",
			log.Ldate|log.Ltime,
		),
	}
}

func (node *Node) newApiServer() *http.Server {
	//mux := http.NewServeMux()
	//mux.HandleFunc("/blocks", node.blocksHandler)
	//mux.HandleFunc("/mineBlock", node.mineBlockHandler)
	//mux.HandleFunc("/peers", node.peersHandler)
	//mux.HandleFunc("/addPeer", node.addPeerHandler)

	yaag.Init(&yaag.Config{On: true, DocTitle: "Gorilla Mux", DocPath: "./apidoc/apidoc.html"})
	router := mux.NewRouter()

	//router.Handle("/apidoc", http.FileServer(http.Dir("./apidoc")))
	router.PathPrefix("/doc/").Handler(http.StripPrefix("/doc/", http.FileServer(http.Dir("./apidoc"))))
	router.HandleFunc("/", middleware.HandleFunc(node.sayHello)).Methods("GET")

	router.HandleFunc("/blockchain/print", middleware.HandleFunc(node.printBlockchain)).Methods("GET")
	router.HandleFunc("/block/{hash}", middleware.HandleFunc(node.getBlock)).Methods("GET")

	router.HandleFunc("/wallet/new", middleware.HandleFunc(node.createWallet)).Methods("POST")
	router.HandleFunc("/wallet/{hash}", middleware.HandleFunc(node.getWallet)).Methods("GET")
	router.HandleFunc("/wallets/list", middleware.HandleFunc(node.listWallet)).Methods("GET")

	router.HandleFunc("/prepare", node.prepare).Methods("POST")
	router.HandleFunc("/send", node.send).Methods("POST")

	return &http.Server{
		Handler: router,
		Addr:    ":" + node.apiAddr,
	}
}

//func (node *Node) newP2PServer() *http.Server {
//	return &http.Server{
//		Handler: websocket.Handler(func(ws *websocket.Conn) {
//			conn := NewConn(ws)
//			node.log("connect to peer:", conn.remoteHost())
//			node.addConn(conn)
//			node.p2pHandler(conn)
//		}),
//		Addr: *p2pAddr,
//	}
//	return
//}

func (node *Node) Run(minerAddress string) {
	apiSrv := node.newApiServer()

	go func() {
		node.log("start HTTP server for API")

		if err := apiSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("apiAddr:", node.apiAddr, "nodeID:", node.nodeID, "nodeADD:", node.nodeADD)

	tcpSrv := NewServer(node, minerAddress)
	go func() {
		node.log("start TCP server")
		tcpSrv.Start()
	}()

	//p2pSrv := node.newP2PServer()
	//go func() {
	//	node.log("start WebSocket server for P2P")
	//	if err := p2pSrv.ListenAndServe(); err != nil {
	//		log.Fatal(err)
	//	}
	//}()

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)
	for {
		s := <-signalCh
		if s == syscall.SIGTERM {
			node.log("stop servers")
			apiSrv.Shutdown(context.Background())
			tcpSrv.Stop()
			//p2pSrv.Shutdown(context.Background())
		}
	}
}

func (node *Node) log(v ...interface{}) {
	node.logger.Println(v)
}

func (node *Node) logError(err error) {
	node.log("[ERROR]", err)
}

func (node *Node) writeResponse(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (node *Node) error(w http.ResponseWriter, err error, message string) {
	node.logError(err)

	b, err := json.Marshal(&ErrorResponse{
		Error: message,
	})
	if err != nil {
		node.logError(err)
	}

	node.writeResponse(w, b)
}
