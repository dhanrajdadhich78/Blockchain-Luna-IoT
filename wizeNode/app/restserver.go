package app

import (
	"log"
	"net"
	"net/http"

	//"github.com/betacraft/yaag/middleware"
	"github.com/betacraft/yaag/yaag"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

// TODO: refactoring - names, funcs
// TODO: add logging & error handling
// TODO: Start and Close rethink
// TODO: what is middleware.HandleFunc() doing here?
// TODO: CORS refactoring
// TODO: refactoring exits from all routines

// RestServer provides HTTP service.
type RestServer struct {
	node *Node
	addr string
	ln   net.Listener
}

// New returns an uninitialized HTTP service.
func NewRestServer(node *Node, addr string) *RestServer {
	return &RestServer{
		node: node,
		addr: addr,
	}
}

// Start starts the service.
func (s *RestServer) Start() error {
	yaag.Init(&yaag.Config{On: true, DocTitle: "Gorilla Mux", DocPath: "./apidoc/apidoc.html"})

	// Get the mux router object
	router := mux.NewRouter().StrictSlash(false)

	//router.Handle("/apidoc", http.FileServer(http.Dir("./apidoc")))
	router.PathPrefix("/doc/").Handler(http.StripPrefix("/doc/", http.FileServer(http.Dir("./apidoc"))))

	// TODO: what is middleware.HandleFunc() doing here?
	//router.HandleFunc("/", middleware.HandleFunc(node.sayHello)).Methods("GET")
	router.HandleFunc("/", s.sayHello)

	// inner usage
	router.HandleFunc("/blockchain/print", s.printBlockchain).Methods("GET")
	router.HandleFunc("/block/{hash}", s.getBlock).Methods("GET")

	router.HandleFunc("/wallet/{hash}", s.getWallet).Methods("GET")

	// send transaction steps: prepare/sign
	router.HandleFunc("/prepare", s.prepare).Methods("POST")
	router.HandleFunc("/sign", s.sign).Methods("POST")

	// DEPRECATED: inner usage
	//router.HandleFunc("/wallet/new", s.deprecatedWalletCreate).Methods("POST")
	//router.HandleFunc("/wallets/list", s.deprecatedWalletsList).Methods("GET")
	//router.HandleFunc("/send", s.deprecatedSend).Methods("POST")

	// TODO: CORS refactoring
	corsHandler := cors.AllowAll().Handler(router)

	// Create a negroni instance
	n := negroni.Classic()
	//n.Use(delay.Middleware{})
	n.UseHandler(corsHandler)

	server := http.Server{
		Handler: n,
		Addr:    ":" + s.node.apiADD,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	// TODO: refactoring exits from all routines
	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Printf("HTTP serve: %s", err)
		}
		//shutdown <- 1
	}()

	return nil
}

// Close closes the service.
func (s *RestServer) Close() {
	log.Println("rest closing")
	s.ln.Close()
	return
}
