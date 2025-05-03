package main

import (
	"flag"
	"log"
	"os"

	"github.com/jennxsierra/dualnet-chat/internal/netutils"
	"github.com/jennxsierra/dualnet-chat/internal/udp/client"
)

func main() {
	// Get computer hostname to use as the default name
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("[error] Failed to retrieve hostname:", err)
	}

	clientName := flag.String("name", hostname, "Name of the client")                            // --name flag
	serverAddr := flag.String("server", "127.0.0.1:4001", "Address of the server to connect to") // --server flag
	flag.Parse()

	// Ensure server address is valid
	port, err := netutils.GetPortFromAddress(*serverAddr)
	if err != nil {
		log.Fatalf("[error] Invalid address format %s: %v\n", *serverAddr, err)
	}
	if !netutils.IsValidPort(port) {
		log.Fatalf("[error] Address %s has invalid port number.\n", *serverAddr)
	}

	// Create client and start chat
	client, err := client.NewClient(*serverAddr, *clientName)
	if err != nil {
		log.Fatalf("[error] Unable to connect to server: %v\n", err)
	}
	client.Start()
}
