package main

import (
	"Heis/pkg/network/localip"
	"Heis/pkg/processPairs"
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	id, port := parseFlags()
	id = generateIDIfEmpty(id)
	backupPort := calculateBackupPort(port)

	// Determine process role using processPairs
	connection, err := processPairs.SetupUDPListener(backupPort)
	if err != nil {
		fmt.Println("Starting as primary...")
		go processPairs.StartPrimaryProcess(id, port, backupPort)
		go processPairs.SpawnBackupProcess(id, port)
	} else {
		fmt.Println("Starting as backup...")
		go processPairs.BackupSetup(id, port, connection, backupPort)
		go processPairs.MonitorAndTakeOver(id, port, connection, backupPort)
	}

	select {}
}

// parseFlags handles command-line flag parsing
func parseFlags() (string, string) {
	var id, port string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()
	return id, port
}

// generateIDIfEmpty creates a unique ID if none provided
func generateIDIfEmpty(id string) string {
	if id != "" {
		return id
	}
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
}

// calculateBackupPort defines the backup communication port so the elevators have a unique port to communicate with its backup
func calculateBackupPort(port string) string {
	// adds a offset to the port
	return strconv.Itoa(atoi(port) + 30210)
}

// atoi safely converts string to int
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0 // Default to 0 if conversion fails
	}
	return i
}
