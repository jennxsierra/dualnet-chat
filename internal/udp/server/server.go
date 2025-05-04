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
	"golang.org/x/time/rate"
)

// ClientInfo stores information about a connected UDP client
type ClientInfo struct {
	Name     string
	Addr     *net.UDPAddr
	Limiter  *rate.Limiter
	LastSeen time.Time
}

// Server stores information about its address and connected clients
type Server struct {
	Addr         string
	Conn         *net.UDPConn
	Clients      map[string]*ClientInfo
	mu           sync.Mutex
	shuttingDown bool
	done         chan struct{}
}

// NewServer creates a new UDP server instance given an address
func NewServer(addr string) *Server {
	return &Server{
		Addr:    addr,
		Clients: make(map[string]*ClientInfo),
		done:    make(chan struct{}),
	}
}

// Start initializes the UDP server and starts listening for client messages
func (s *Server) Start() error {
	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}

	// Create a UDP connection
	s.Conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer s.Conn.Close()

	s.monitorTermSig()         // Monitor for termination signal
	s.monitorInactiveClients() // Monitor for inactive clients

	// Welcome message
	fmt.Println("[dualnet-chat UDP Server]")
	log.Printf("[info] Server is listening on %s\n\n", netutils.GetIPv4Addr("udp", udpAddr.Port))

	// Process incoming messages
	return s.processMessages()
}

// processMessages handles incoming UDP messages
func (s *Server) processMessages() error {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-s.done:
			return nil
		default:
			// Set a read deadline so we can check for server shutdown
			s.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			// Read message
			n, addr, err := s.Conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is just a timeout from our deadline, not a real error
					continue
				}
				log.Printf("[error] Reading from UDP: %v", err)
				continue
			}

			// Reset read deadline
			s.Conn.SetReadDeadline(time.Time{})

			// Process the message
			message := strings.TrimSpace(string(buffer[:n]))
			s.handleMessage(addr, message)
		}
	}
}

// handleMessage processes a single message from a UDP client
func (s *Server) handleMessage(addr *net.UDPAddr, message string) {
	addrStr := addr.String()

	// Check if this is a new client (registration message)
	s.mu.Lock()
	client, exists := s.Clients[addrStr]
	s.mu.Unlock()

	if !exists {
		// This is a new client, register them
		if strings.HasPrefix(message, "REGISTER:") {
			clientName := strings.TrimPrefix(message, "REGISTER:")
			s.registerClient(addr, clientName)
		}
		return
	}

	// Update last seen time for existing client
	s.mu.Lock()
	client.LastSeen = time.Now()
	s.mu.Unlock()

	// Handle heartbeat message
	if message == "HEARTBEAT" {
		return
	}

	// Handle disconnect message
	if message == "BYE" {
		s.handleClientDisconnect(addr)
		return
	}

	// Handle regular message
	// Check rate limit
	if client.Limiter.Allow() {
		// Broadcast message to all other clients
		s.broadcast(fmt.Sprintf("[%s]: %s", client.Name, message), addr)
	} else {
		// Notify client they're sending too fast
		s.Conn.WriteToUDP([]byte("[server]: You are sending messages too fast. Please slow down.\n"), addr)
	}
}

// handleClientDisconnect processes a client disconnection
func (s *Server) handleClientDisconnect(addr *net.UDPAddr) {
	addrStr := addr.String()

	s.mu.Lock()
	client, exists := s.Clients[addrStr]
	if exists {
		clientName := client.Name
		delete(s.Clients, addrStr)
		s.mu.Unlock()

		// Log the disconnection
		log.Printf("[-] %s", clientName)
		// Broadcast disconnection message to other clients
		s.broadcast(fmt.Sprintf("[-] %s left the chat\n", clientName), addr)
	} else {
		s.mu.Unlock()
	}
}

// registerClient adds a new client to the server
func (s *Server) registerClient(addr *net.UDPAddr, name string) {
	// Create a limiter for this client: 1 message per second with burst of 3
	limiter := rate.NewLimiter(1, 3)

	s.mu.Lock()
	s.Clients[addr.String()] = &ClientInfo{
		Name:     name,
		Addr:     addr,
		Limiter:  limiter,
		LastSeen: time.Now(),
	}
	s.mu.Unlock()

	// Log and broadcast client connection
	log.Printf("[+] %s", name)
	s.broadcast(fmt.Sprintf("[+] %s joined the chat\n", name), addr)

	// Send confirmation to the client
	s.Conn.WriteToUDP([]byte(fmt.Sprintf("[server]: Welcome %s, you are now registered!\n\n", name)), addr)
}

// broadcast sends a message to all connected clients except the sender
func (s *Server) broadcast(message string, ignoreAddr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.Clients {
		if client.Addr.String() != ignoreAddr.String() {
			s.Conn.WriteToUDP([]byte(message+"\n"), client.Addr)
		}
	}
}

// monitorInactiveClients periodically checks for clients that haven't sent messages recently
func (s *Server) monitorInactiveClients() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				inactiveThreshold := 2 * time.Minute

				s.mu.Lock()
				for addrStr, client := range s.Clients {
					if now.Sub(client.LastSeen) > inactiveThreshold {
						// Client hasn't sent a message in too long, consider them disconnected
						delete(s.Clients, addrStr)
						log.Printf("[-] %s (inactive timeout)", client.Name)
						s.broadcast(fmt.Sprintf("[-] %s left the chat (timeout)\n", client.Name), client.Addr)
					}
				}
				s.mu.Unlock()
			case <-s.done:
				return
			}
		}
	}()
}

// monitorTermSig listens for termination signals and gracefully shuts down
func (s *Server) monitorTermSig() {
	signalChan := make(chan os.Signal, 1)

	// Register channel to receive interrupt and termination OS signals
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// When a proper signal is received, disconnect all clients and exit program
	go func() {
		<-signalChan
		fmt.Println() // Print a newline for neatness
		log.Println("[info] Server shutting down...")

		// Set shutting down flag and notify all clients
		s.mu.Lock()
		s.shuttingDown = true
		for _, client := range s.Clients {
			log.Printf("[-] Disconnecting %s", client.Name)
			s.Conn.WriteToUDP([]byte("[server]: Server is shutting down. Goodbye!\n"), client.Addr)
		}
		s.mu.Unlock()

		// Signal to stop all goroutines
		close(s.done)

		os.Exit(0)
	}()
}
