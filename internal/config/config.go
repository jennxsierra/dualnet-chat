package config

import (
	"flag"
	"fmt"
	"os"
)

// LoadConfig process and validates a "port" flag, whose default is set by its parameter.
func LoadConfig(defaultPort int) Config {
	port := flag.Int("port", defaultPort, "Port to run the application on")
	flag.Parse()

	if !validatePort(*port) {
		fmt.Fprintf(os.Stderr, "[ERROR] Port %d is invalid. Port must be between 1 and 65535.\n", *port)
		os.Exit(1)
	}

	return Config{
		Port: *port,
	}
}

// Config stores the value of the application port.
type Config struct {
	Port int
}

// validatePort returns true if passed in port is with the valid range and false otherwise.
func validatePort(port int) bool {
	return port >= 1 && port <= 65535
}
