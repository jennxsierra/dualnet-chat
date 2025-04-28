package client

import (
	"fmt"
	"net"
	"strings"

	"github.com/chzyer/readline"
	"github.com/jennxsierra/dualnet-chat/internal/netutils"
)

// Client stores the client connection and name.
type Client struct {
	Conn net.Conn
	Name string
	rl   *readline.Instance
	done chan struct{}
}

// NewClient creates a new client instance that connects to the server.
func NewClient(serverAddr string, name string) (*Client, error) {
	// establish the TCP connection
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	// get the connection's IPv4 address and port
	clientAddr := netutils.GetIPv4Addr("tcp", conn.LocalAddr().(*net.TCPAddr).Port)
	fullName := fmt.Sprintf("%s@%s", name, clientAddr) // e.g. AHARCH@192.168.18.4:51756

	// create readline instance
	rl, err := readline.New(fmt.Sprintf("[%s]: ", fmt.Sprintf("%s@%s", name, clientAddr)))
	if err != nil {
		return nil, err
	}

	client := &Client{
		Conn: conn,
		Name: fullName,
		rl:   rl,
		done: make(chan struct{}),
	}

	return client, nil
}

// Start starts the client, connecting to the server and handling messages
func (c *Client) Start() {
	// welcome message
	fmt.Println("[dualnet-chat TCP Client]")
	fmt.Printf("[info] You are connected to [%s] as [%s]\n\n", c.Conn.RemoteAddr(), c.Name)

	// send the name as the first message to the server
	fmt.Fprintf(c.Conn, "%s\n", c.Name)

	go c.handleMessages()
	c.sendMessages()
}

// handleMessages listens for messages from the server and prints them to the console.
func (c *Client) handleMessages() {
	scanner := readline.NewCancelableStdin(c.Conn)
	buf := make([]byte, 4096)
	for {
		// read server message
		n, err := scanner.Read(buf)
		if err != nil {
			break
		}
		message := string(buf[:n])

		// print server message and refresh screen
		c.rl.Write([]byte(message))
		c.rl.Refresh()
	}

	// if cannot read from server because the server disconnected,
	// notify channel and close readline
	close(c.done)
	c.rl.Close()
}

// sendMessages reads user input from standard input and sends it to the server.
func (c *Client) sendMessages() {
	// continuously reading user input
	for {
		line, err := c.rl.Readline()

		// return on error
		if err != nil {
			select {
			case <-c.done: // check if server was disconnected
				fmt.Println("\n[info] Server disconnected. Exiting...")
			default: // user disconnects themselves
				fmt.Println("\nGoodbye!")
			}

			return
		}

		// trim whitespace
		line = strings.TrimSpace(line)

		// send message to the server
		fmt.Fprintf(c.Conn, "%s\n", line)
	}
}
