package client

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/jennxsierra/dualnet-chat/internal/netutils"
)

// Client stores the UDP client connection and details
type Client struct {
	conn       *net.UDPConn
	serverAddr *net.UDPAddr
	Name       string
	rl         *readline.Instance
	done       chan struct{}
}

// NewClient creates a new UDP client that connects to the server
func NewClient(serverAddr string, name string) (*Client, error) {
	// Resolve server address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	// Get the client's local address
	clientAddr := netutils.GetIPv4Addr("udp", conn.LocalAddr().(*net.UDPAddr).Port)
	fullName := fmt.Sprintf("%s@%s", name, clientAddr)

	// Create readline instance
	rl, err := readline.New(color.YellowString("[%s]: ", fullName))
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:       conn,
		serverAddr: udpAddr,
		Name:       fullName,
		rl:         rl,
		done:       make(chan struct{}),
	}

	return client, nil
}

// Start initializes the UDP client and starts handling messages
func (c *Client) Start() {
	// Welcome message
	fmt.Println("[dualnet-chat UDP Client]")
	fmt.Printf("[info] You are connected to [%s] as [%s]\n\n", c.serverAddr, c.Name)

	// Register with the server
	c.register()

	// Start heartbeat to keep the connection active
	go c.sendHeartbeats()

	// Start handling messages
	go c.handleMessages()

	// Start sending user messages
	c.sendMessages()
}

// register sends the client's name to the server for registration
func (c *Client) register() {
	// Send registration message
	_, err := c.conn.Write([]byte(fmt.Sprintf("REGISTER:%s", c.Name)))
	if err != nil {
		fmt.Printf("[error] Failed to register with server: %v\n", err)
		return
	}
}

// sendHeartbeats periodically sends heartbeat messages to keep the connection active
func (c *Client) sendHeartbeats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.conn.Write([]byte("HEARTBEAT"))
		case <-c.done:
			return
		}
	}
}

// handleMessages listens for messages from the server and displays them
func (c *Client) handleMessages() {
	buffer := make([]byte, 4096)

	for {
		// Set read deadline to check for done channel periodically
		c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, _, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// This is just our timeout, not a real error
				select {
				case <-c.done:
					return
				default:
					continue
				}
			}

			// Real error, likely server disconnected
			close(c.done)
			c.rl.Write([]byte("\n[info] Connection to server lost. Exiting...\n"))
			c.rl.Close()
			return
		}

		// Reset read deadline
		c.conn.SetReadDeadline(time.Time{})

		message := string(buffer[:n])

		// Print server message and refresh screen
		c.rl.Write([]byte(message))
		c.rl.Refresh()
	}
}

// sendMessages reads user input and sends it to the server
func (c *Client) sendMessages() {
	// Continuously reading user input
	for {
		line, err := c.rl.Readline()

		// Return on error
		if err != nil {
			select {
			case <-c.done: // Check if server was disconnected
				fmt.Println("\n[info] Server disconnected. Exiting...")
			default: // User disconnects themselves
				fmt.Println("\nGoodbye!")
				close(c.done)
			}

			return
		}

		// Trim whitespace
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Send message to the server
		_, err = c.conn.Write([]byte(line))
		if err != nil {
			c.rl.Write([]byte(fmt.Sprintf("[error] Failed to send message: %v\n", err)))
			c.rl.Refresh()
		}
	}
}
