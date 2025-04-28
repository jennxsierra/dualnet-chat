package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jennxsierra/dualnet-chat/internal/tcp/client"
)

// Server stores information about its address and connected clients.
type Server struct {
	Addr    string
	Clients map[net.Conn]*client.Client
	mu      sync.Mutex
}

// NewServer creates a [Server] instance given an address.
func NewServer(addr string) *Server {
	return &Server{
		Addr:    addr,
		Clients: make(map[net.Conn]*client.Client),
	}
}

// Start listens on the established address and launches a goroutine for every
// successfully connected client.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.monitorTermSig() // monitor for termination signal

	// welcome message
	fmt.Println("[dualnet-chat TCP Server]")
	log.Printf("[info] Server is listening on %s\n\n", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[error]", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection reads client messages and broadcasts it to all connected clients.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		clientName := scanner.Text() // read the client's name first

		// add new client to the server map
		s.mu.Lock()
		c := &client.Client{Conn: conn, Name: clientName}
		s.Clients[conn] = c
		s.mu.Unlock()

		// log and broadcast client connection
		log.Printf("[+] %s", c.Name)
		s.broadcast(fmt.Sprintf("[+] %s joined the chat\n", c.Name), conn)

		// continuously read and broadcast client message until disconnect
		for scanner.Scan() {
			text := scanner.Text()
			s.broadcast(fmt.Sprintf("[%s]: %s\n", c.Name, text), conn)
		}

		// remove disconnected client from the server map
		s.mu.Lock()
		delete(s.Clients, conn)
		s.mu.Unlock()

		// log and broadcast client disconnection
		log.Printf("[-] %s", c.Name)
		s.broadcast(fmt.Sprintf("[-] %s left the chat\n", c.Name), conn)
	}
}

// broadcast sends a client message to all clients in the server map
// except the sending client.
func (s *Server) broadcast(message string, ignoreConn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// write message to every connected client
	for conn, c := range s.Clients {
		if conn != ignoreConn {
			fmt.Fprint(c.Conn, message)
		}
	}
}

// monitorTermSig listens for a termination signal, and upon receiving one,
// prints a message and disconnects every connected client.
func (s *Server) monitorTermSig() {
	signalChan := make(chan os.Signal, 1)

	// register channel to receive interrupt and termination OS signals
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// when a proper signal is received, disconnect all clients and exit program
	go func() {
		<-signalChan
		fmt.Println() // print a newline for neatness
		log.Println("[info] Server shutting down...")

		// close all client connections
		s.mu.Lock()
		for conn, c := range s.Clients {
			log.Printf("[-] Disconnecting %s", c.Name)
			conn.Close()
		}
		s.mu.Unlock()

		os.Exit(0)
	}()
}
