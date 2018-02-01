package network

import "flag"

var (
	apiAddr   = flag.String("api", ":3001", "HTTP server address for API")
	p2pAddr   = flag.String("p2p", ":6001", "WebSocket server address for P2P")
	p2pOrigin = flag.String("origin", "http://127.0.0.1", "P2P origin")
)
