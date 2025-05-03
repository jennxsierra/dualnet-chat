package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jennxsierra/dualnet-chat/internal/netutils"
	"github.com/jennxsierra/dualnet-chat/internal/udp/server"
)

func main() {
	port := flag.Int("port", 4001, "Port to run the UDP server on") // --port flag
	flag.Parse()

	// Ensure port is within the valid range
	if !netutils.IsValidPort(*port) {
		log.Fatalf("[error] Port %d is invalid. Port must be between 1 and 65535.\n", *port)
	}

	// Create and start server
	server := server.NewServer(fmt.Sprintf("0.0.0.0:%d", *port))
	if err := server.Start(); err != nil {
		log.Fatalf("[error] Server failed to start: %v\n", err)
	}
}
