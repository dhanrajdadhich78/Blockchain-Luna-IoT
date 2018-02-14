package app

import (
	"github.com/gorilla/mux"
	"net/http"

	"log"
	b "wizeBlockchain/blockchain"
)

type ApiHandler struct {
	bc *b.Blockchain
}

func startApiServer(apiAddr string, bc *b.Blockchain) {
	router := mux.NewRouter()
	router.HandleFunc("/", getHello).Methods("GET")
	//// transactions
	//a.Router.HandleFunc("/transaction", a.newTransaction).Methods("POST")
	//a.Router.HandleFunc("/transaction/distributed", a.distributedTransaction).Methods("POST")
	//a.Router.HandleFunc("/transactions/{hash}", a.transactions).Methods("GET")
	//a.Router.HandleFunc("/transactions", a.currentTransactions).Methods("GET")
	// wallet
	//router.HandleFunc("/wallet/{hash}", getWallet).Methods("GET")
	//// blocks
	//a.Router.HandleFunc("/block", a.lastblock).Methods("GET")
	//a.Router.HandleFunc("/block/{hash}", a.block).Methods("GET")
	//a.Router.HandleFunc("/block/index/{index}", a.blockByIndex).Methods("GET")
	//a.Router.HandleFunc("/block/distributed", a.distributedBlock).Methods("POST")
	//// mining and chaining
	//a.Router.HandleFunc("/mine", a.mine).Methods("GET")
	//a.Router.HandleFunc("/chain", a.chain).Methods("GET")
	//a.Router.HandleFunc("/validate", a.validate).Methods("GET")
	//a.Router.HandleFunc("/resolve", a.resolve).Methods("GET")
	//a.Router.HandleFunc("/status", a.chainStatus).Methods("GET")
	//// Clients
	//a.Router.HandleFunc("/client", a.connectClient).Methods("POST")
	//a.Router.HandleFunc("/client", a.getClients).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+apiAddr, router))

}

func getHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize")
}

//func getWallet(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	hash := vars["hash"]
//	nodeID := os.Getenv("NODE_ID")
//	fmt.Println(hash)
//	resp := map[string]interface{}{
//		"success": true,
//		"credit":  GetWalletCredits(hash, nodeID),
//	}
//	respondWithJSON(w, http.StatusOK, resp)
//}
