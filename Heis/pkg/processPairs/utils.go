package processPairs

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"
)

// atoi safely converts string to int
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0 // Default to 0 if conversion fails
	}
	return i
}

// StartElevatorSystem determines the role (primary/backup) and starts accordingly
func StartElevatorSystem(cfg *Config) error {
	addr := ":" + cfg.BackupPort
	s, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("error resolving UDP address: %v", err)
	}

	// Attempt to listen on backup port to determine role
	connection, err := net.ListenUDP("udp", s)
	if err != nil {
		fmt.Println("Starting as primary...")
		go RunPrimary(cfg)
		go spawnInitialBackup(cfg)
	} else {
		fmt.Println("Starting as backup...")
		go RunBackup(cfg, connection)
	}
	return nil
}

// spawnInitialBackup spawns the initial backup process
func spawnInitialBackup(cfg *Config) {
	time.Sleep(1 * time.Second)
	err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
		"-id", cfg.ID+"-backup", "-port", cfg.Port).Run()
	if err != nil {
		log.Printf("Failed to spawn backup: %v", err)
	}
}
