package httpd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var StartTime time.Time

type Ping struct {
	Ping string
}
type Pong struct {
	FreeStorage uint64
	Uptime      int64
}

func getOnlyHash(data string) string {
	h256 := sha256.New()
	out := fmt.Sprintf("%s", data)
	io.WriteString(h256, out)

	return fmt.Sprintf("%x", h256.Sum(nil))
}

func (s *Service) handleEchoRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		ping := Ping{}
		pong := Pong{}
		err := json.NewDecoder(r.Body).Decode(&ping)
		if err != nil {
			respondWithJSON(w, http.StatusBadRequest, "Bad request")
		}

		if ping.Ping == getOnlyHash("pingraft") {
			now := time.Now()
			pong.Uptime = int64(now.Sub(StartTime).Seconds())
			pong.FreeStorage = 0

			respondWithJSON(w, http.StatusOK, pong)
		} else {
			respondWithJSON(w, http.StatusForbidden, "auth required")
		}

	} else {
		respondWithJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
	return
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
