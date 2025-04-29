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
var testLogger *log.Logger
// log file path
const logFilePath = "results/tcp_tests.log"
func init() {
	// ensure the log directory exists
	logDir := "results"
	err := os.MkdirAll(logDir, 0755) // create the directory if it doesn't exist
	if err != nil {
		log.Printf("Warning: Could not create log directory %s: %v", logDir, err)
	}

	var writers []io.Writer = []io.Writer{os.Stdout}

	// open the log file for writing
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		writers = append(writers, file)
	} else {
		log.Printf("Warning: Could not log to file %s: %v", logFilePath, err)
	}

	// initialize the logger with the writers
	testLogger = log.New(io.MultiWriter(writers...), "", log.LstdFlags)
}

// TestConnectionLatency tests for the roundtrip time of a packet sent to and from
// the server. In TCP connections, the client sends a SYN, receives a SYNACK from the
// server, then sends an ACK to the server.
//
// [net.Dial] completes when SYNACK is received, so the measured latency can be expected
// to be approximately double that of what was artificially induced with tc.
func TestConnectionLatency(t *testing.T) {
	const serverAddr = "127.0.0.1:4000"

	start := time.Now()
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	latency := time.Since(start)
	defer conn.Close()

	testLogger.Printf("Measured TCP connection latency: %v\n", latency)
}

// TestThroughput measures how long to send a 5MB payload to the server in 4KB
// chunks. The result is in MB/s and the result will vary depending the the
// impaired network conditions set by tc in the Makefile.
func TestThroughput(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:4000")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// send a dummy name
	clientName := "TestThroughput"
	_, err = conn.Write([]byte(clientName))
	if err != nil {
		t.Fatalf("Failed to send client name: %v", err)
	}

	payload := make([]byte, 1024*1024*5) // 5MB total
	chunkSize := 4096                    // 4KB chunk

	// measure the time it takes to write the total length of the payload
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
	}
	duration := time.Since(start)

	testLogger.Printf("Sent %d bytes in %v (%.2f MB/s)\n", totalWritten, duration, float64(totalWritten)/(1024*1024)/duration.Seconds())
}
