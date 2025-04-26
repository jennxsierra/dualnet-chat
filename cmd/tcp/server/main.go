package main

import (
	"fmt"

	"github.com/jennxsierra/dualnet-chat/internal/config"
)

func main() {
	cfg := config.LoadConfig(4000)
	fmt.Printf("Port: %v\n", cfg.Port)
}
