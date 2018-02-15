package app

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	ww "wizeBlock/wizeNode/wallet"
)

func (node *Node) sayHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize "+node.nodeADD)
}

func (node *Node) getWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	resp := map[string]interface{}{
		"success": true,
		"credit":  GetWalletCredits(hash, node.nodeID, node.blockchain),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) listWallet(w http.ResponseWriter, r *http.Request) {
	wallets, err := ww.NewWallets(node.nodeID)
	if err != nil {
		log.Panic(err)
	}

	resp := map[string]interface{}{
		"success":     true,
		"listWallets": wallets.GetAddresses(),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
