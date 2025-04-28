package network

import (
	"net"
	"testing"
	"time"
)

func TestConnectionLatency(t *testing.T) {
	const serverAddr = "127.0.0.1:4000"

	start := time.Now()
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	latency := time.Since(start)
	defer conn.Close()

	t.Logf("Measured TCP connection latency: %v (approx 2× one-way network delay)", latency)
}

func TestThroughput(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:4000")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// send a name first
	clientName := "TestThroughput"
	_, err = conn.Write([]byte(clientName))
	if err != nil {
		t.Fatalf("Failed to send client name: %v", err)
	}

	// now prepare the payload
	payload := make([]byte, 1024*1024*5) // 5MB payload

	start := time.Now()

	totalWritten := 0
	chunkSize := 4096 // 4KB per chunk
	for totalWritten < len(payload) {
		end := totalWritten + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		n, err := conn.Write(payload[totalWritten:end])
		if err != nil {
			t.Fatalf("Failed during payload send: %v", err)
		}
		totalWritten += n
	}

	duration := time.Since(start)
	t.Logf("Sent %d bytes in %v (%.2f MB/s)", totalWritten, duration, float64(totalWritten)/(1024*1024)/duration.Seconds())
}
