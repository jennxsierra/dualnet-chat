package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jennxsierra/dualnet-chat/internal/netutils"
)

// Client stores the client connection and name.
type Client struct {
	Conn net.Conn
	Name string
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

	// send the name as the first message to the server
	fmt.Fprintf(c.Conn, "%s\n", c.Name)

	go c.handleMessages()
	c.sendMessages()
}

// handleMessages listens for messages from the server and prints them to the console.
func (c *Client) handleMessages() {
	scanner := bufio.NewScanner(c.Conn)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

// sendMessages reads user input from standard input and sends it to the server.
func (c *Client) sendMessages() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("[%s]: ", c.Name)
		scanner.Scan()
		message := scanner.Text()

		// close the connection on "/exit"
		if strings.ToLower(message) == "/exit" {
			fmt.Println("Exiting...")
			c.Conn.Close()
			break
		}

		fmt.Fprintf(c.Conn, "%s\n", message)
	}
}
