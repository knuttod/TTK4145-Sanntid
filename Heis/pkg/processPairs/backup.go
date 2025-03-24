package processPairs

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"
)

// RunBackup runs the elevator system as the backup process
func RunBackup(cfg *Config, connection *net.UDPConn) {
	defer connection.Close()

	buffer := make([]byte, 1024)

	for {
		connection.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := connection.ReadFromUDP(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Primary failed, taking over as primary...")
				go spawnNewBackup(cfg)
				go RunPrimary(cfg)
				return
			}
		}
	}
}

// spawnNewBackup starts a new backup process
func spawnNewBackup(cfg *Config) {
	err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
		"-id", cfg.ID+"-backup", "-port", cfg.Port).Run()
	if err != nil {
		log.Printf("Failed to spawn new backup: %v", err)
	}
}
