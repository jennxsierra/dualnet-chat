package network

import (
	"io"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

// logger for printing to standard output and a log file
var udpTestLogger *log.Logger

// log file path
const udpLogFilePath = "results/udp_tests.log"

func init() {
	// ensure the log directory exists
	logDir := "results"
	err := os.MkdirAll(logDir, 0755) // create the directory if it doesn't exist
	if err != nil {
		log.Printf("Warning: Could not create log directory %s: %v", logDir, err)
	}

	var writers []io.Writer = []io.Writer{os.Stdout}

	// open the log file for writing
	file, err := os.OpenFile(udpLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		writers = append(writers, file)
	} else {
		log.Printf("Warning: Could not log to file %s: %v", udpLogFilePath, err)
	}

	// initialize the logger with the writers
	udpTestLogger = log.New(io.MultiWriter(writers...), "", log.LstdFlags)
}

// TestUDPLatency measures round-trip time for UDP packets
func TestUDPLatency(t *testing.T) {
	const serverAddr = "127.0.0.1:4001"

	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		t.Fatalf("Failed to connect to UDP server: %v", err)
	}
	defer conn.Close()

	// Register with server
	clientName := "TestUDPLatency"
	start := time.Now()
	_, err = conn.Write([]byte("REGISTER:" + clientName))
	if err != nil {
		t.Fatalf("Failed to register with server: %v", err)
	}

	// Wait for welcome message
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Failed to receive welcome message: %v", err)
	}

	latency := time.Since(start)
	udpTestLogger.Printf("Measured UDP round-trip latency: %v\n", latency)
}

// TestUDPThroughput measures how long it takes to send a payload to the server
func TestUDPThroughput(t *testing.T) {
	const serverAddr = "127.0.0.1:4001"

	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		t.Fatalf("Failed to connect to UDP server: %v", err)
	}
	defer conn.Close()

	// Register with server
	clientName := "TestUDPThroughput"
	_, err = conn.Write([]byte("REGISTER:" + clientName))
	if err != nil {
		t.Fatalf("Failed to register with server: %v", err)
	}

	// Wait for welcome message
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Failed to receive welcome message: %v", err)
	}

	// Smaller payload for UDP to avoid fragmentation issues
	payload := make([]byte, 1024*512) // 512KB total
	chunkSize := 1024                 // 1KB chunk (UDP packet size should be kept small)

	// Measure the time to write the payload
	totalWritten := 0
	start := time.Now()
	for totalWritten < len(payload) {
		chunk := payload[totalWritten:]
		if len(chunk) > chunkSize {
			chunk = chunk[:chunkSize]
		}
		n, err := conn.Write(chunk)
		if err != nil {
			t.Fatalf("Failed during payload send: %v", err)
		}
		totalWritten += n

		// Small delay to avoid overwhelming the network
		time.Sleep(time.Millisecond)
	}
	duration := time.Since(start)

	udpTestLogger.Printf("Sent %d bytes in %v (%.2f KB/s)\n",
		totalWritten, duration, float64(totalWritten)/1024/duration.Seconds())
}
