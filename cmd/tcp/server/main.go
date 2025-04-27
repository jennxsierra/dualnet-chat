package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jennxsierra/dualnet-chat/internal/netutils"
	"github.com/jennxsierra/dualnet-chat/internal/tcp/server"
)

func main() {
	port := flag.Int("port", 4000, "Port to run the TCP server on") // --port flag
	flag.Parse()

	// ensure port is within the valid range
	if !netutils.IsValidPort(*port) {
		log.Fatalf("[error] Port %d is invalid. Port must be between 1 and 65535.\n", *port)
	}

	// create and start server
	server := server.NewServer(fmt.Sprintf("0.0.0.0:%d", *port))
	server.Start()
}
