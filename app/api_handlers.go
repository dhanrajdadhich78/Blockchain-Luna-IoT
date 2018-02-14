package app

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

func (node *Node) sayHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize "+node.apiAddr)
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

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
