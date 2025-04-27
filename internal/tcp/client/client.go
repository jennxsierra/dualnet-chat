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
}

// NewClient creates a new client instance that connects to the server.
func NewClient(serverAddr string, name string) (*Client, error) {
	// establish the TCP connection
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	clientAddr := netutils.GetIPv4Addr("tcp", conn.LocalAddr().(*net.TCPAddr).Port)
	client := &Client{
		Conn: conn,
		Name: fmt.Sprintf("%s@%s", name, clientAddr), // e.g. AHARCH@192.168.18.4:51756
	}

	return client, nil
}

// Start starts the client, connecting to the server and handling messages
func (c *Client) Start() {
	// welcome message
	fmt.Println("[dualnet-chat TCP Client]")
	fmt.Println("[info] Enter \"/exit\" to disconnect")
	fmt.Printf("[info] You are connected to [%s] as [%s]\n\n", c.Conn.RemoteAddr(), c.Name)

	// initialize readline
	rl, err := readline.New(fmt.Sprintf("[%s]: ", c.Name))
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	c.rl = rl

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
}

// sendMessages reads user input from standard input and sends it to the server.
func (c *Client) sendMessages() {
	for {
		// read client message
		line, err := c.rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		// close the connection on "/exit"
		if strings.ToLower(line) == "/exit" {
			fmt.Println("\nGoodbye!")
			c.Conn.Close()
			break
		}

		// send message to the server
		fmt.Fprintf(c.Conn, "%s\n", line)
	}
}
