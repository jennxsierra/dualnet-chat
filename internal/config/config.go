package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Port int
}

func LoadConfig() Config {
	port := flag.Int("port", 4000, "Port to run the application on")
	flag.Parse()

	if !validatePort(*port) {
		fmt.Fprintf(os.Stderr, "[ERROR] Port %d is invalid. Port must be between 1 and 65535.\n", *port)
		os.Exit(1)
	}

	return Config{
		Port: *port,
	}
}

// validatePort returns true if passed in port is with the valid range and false otherwise.
func validatePort(port int) bool {
	return port >= 1 && port <= 65535
}
