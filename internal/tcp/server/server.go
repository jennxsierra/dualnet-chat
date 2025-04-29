package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jennxsierra/dualnet-chat/internal/netutils"
	"github.com/jennxsierra/dualnet-chat/internal/tcp/client"
	"golang.org/x/time/rate"
)

// ServerClient wraps [client.Client] along with a rate limiter.
type ServerClient struct {
	Client  *client.Client
	Limiter *rate.Limiter
}

// Server stores information about its address and connected clients.
type Server struct {
	Addr         string
	Clients      map[net.Conn]*ServerClient
	mu           sync.Mutex
	shuttingDown bool
}

// NewServer creates a [Server] instance given an address.
func NewServer(addr string) *Server {
	return &Server{
		Addr:    addr,
		Clients: make(map[net.Conn]*ServerClient),
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
	log.Printf("[info] Server is listening on %s\n\n", netutils.GetIPv4Addr("tcp", listener.Addr().(*net.TCPAddr).Port))

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

	// periodially check TCP connection for sudden client disconnects (e.g. closing terminal window)
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second) // shorter than default
	}

	// read the client's name first
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if !s.shuttingDown && err.Error() != "EOF" { // do not print if shutting down
			log.Println("Error reading client name:", err)
		}
		return
	}
	clientName := strings.TrimSpace(string(buf[:n]))

	// create a limiter for this client: 1 message per second with burst of 3
	limiter := rate.NewLimiter(1, 3)

	// add the new client to the server map
	s.mu.Lock()
	c := &ServerClient{
		Client:  &client.Client{Conn: conn, Name: clientName},
		Limiter: limiter,
	}
	s.Clients[conn] = c
	s.mu.Unlock()

	// log and broadcast client connection
	log.Printf("[+] %s", c.Client.Name)
	s.broadcast(fmt.Sprintf("[+] %s joined the chat\n", c.Client.Name), conn)

	// continuously read and broadcast client messages until disconnect
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if !s.shuttingDown && err.Error() != "EOF" { // do not print if shutting down
				log.Println("Error reading client message:", err)
			}
			break
		}

		// don't broadcast empty messages
		if n > 0 {
			// retrieve client
			s.mu.Lock()
			sc := s.Clients[conn]
			s.mu.Unlock()

			// client may have disconnected
			if sc == nil {
				break
			}

			// check if the client is allowed to send the message
			if sc.Limiter.Allow() {
				text := string(buf[:n])
				s.broadcast(fmt.Sprintf("[%s]: %s", sc.Client.Name, text), conn)
			} else {
				// optionally notify user they're sending too fast
				fmt.Fprint(conn, "[server]: You are sending messages too fast. Please slow down.\n")
			}
		}
	}

	// remove disconnected client from the server map
	s.mu.Lock()
	delete(s.Clients, conn)
	s.mu.Unlock()

	// log and broadcast client disconnection
	log.Printf("[-] %s", c.Client.Name)
	s.broadcast(fmt.Sprintf("[-] %s left the chat\n", c.Client.Name), conn)
}

// broadcast sends a client message to all clients in the server map
// except the sending client.
func (s *Server) broadcast(message string, ignoreConn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// write message to every connected client
	for conn, sc := range s.Clients {
		if conn != ignoreConn {
			fmt.Fprint(sc.Client.Conn, message)
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

		// close all client connections and set shuttingDown to true
		s.mu.Lock()
		s.shuttingDown = true
		for conn, c := range s.Clients {
			log.Printf("[-] Disconnecting %s", c.Client.Name)
			conn.Close()
		}
		s.mu.Unlock()

		os.Exit(0)
	}()
}
