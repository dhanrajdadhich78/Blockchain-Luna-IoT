package app

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func startApiServer(apiAddr string) {
	router := mux.NewRouter()
	router.HandleFunc("/", getHello).Methods("GET")
	//router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	//router.HandleFunc("/people/{id}", CreatePerson).Methods("POST")
	//router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":"+apiAddr, router))

}

func getHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize")
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
