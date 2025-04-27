package main

import (
	"flag"
	"log"
	"os"

	"github.com/jennxsierra/dualnet-chat/internal/netutils"
	"github.com/jennxsierra/dualnet-chat/internal/tcp/client"
)

func main() {
	// get computer hostname to use as the default name
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("[error] Failed to retrieve hostname:", err)
	}

	clientName := flag.String("name", hostname, "Name of the client")                            // --name flag
	serverAddr := flag.String("server", "127.0.0.1:4000", "Address of the server to connect to") // --server flag
	flag.Parse()

	// ensure server address is valid
	if !netutils.IsValidTCPAddress(*serverAddr) {
		log.Fatalf("[error] Address %s is invalid.\n", *serverAddr)
	}

	// create client and start chat
	client, err := client.NewClient(*serverAddr, *clientName)
	if err != nil {
		log.Fatalf("[error] Unable to connect to server: %v\n", err)
	}
	client.Start()
}
