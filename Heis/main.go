package main

import (
	"Heis/pkg/network/localip"
	"Heis/pkg/processPairs"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	var id string
	var port string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	// Runtime configuration
	cfg := NewConfig(id, port)

	// Start the system
	err := processPairs.StartElevatorSystem(cfg)
	if err != nil {
		log.Fatalf("Failed to start elevator system: %v", err)
	}

	// Keep main running
	select {}
}

// Config holds runtime configuration
type Config struct {
	ID         string
	Port       string
	BackupPort string
}

// NewConfig initializes a Config struct with default values if necessary
func NewConfig(id, port string) *Config {
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	backupPort := strconv.Itoa(processPairs.atoi(port) + 30210)
	return &Config{ID: id, Port: port, BackupPort: backupPort}
}
