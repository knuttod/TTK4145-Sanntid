package main

import (
	"Heis/pkg/network/localip"
	"Heis/pkg/processPairs"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	id, port := parseFlags()
	id = generateIDIfEmpty(id)
	backupPort := calculateBackupPort(port)

	// Setup UDP listener to determine process role
	connection, err := setupUDPListener(backupPort)
	if err != nil {
		fmt.Println("Starting as primary...")
		go startPrimaryProcess(id, port, backupPort)
		go spawnBackupProcess(id, port)
	} else {
		fmt.Println("Starting as backup...")
		go processPairs.BackupSetup(id, port, connection, backupPort)
		go monitorAndTakeOver(id, port, connection, backupPort)
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

// calculateBackupPort defines the backup communication port so all elevators communicate with its backup on different ports.
func calculateBackupPort(port string) string {
	return strconv.Itoa(atoi(port) + 30210)
}

// setupUDPListener attempts to create a UDP listener to determine role
func setupUDPListener(backupPort string) (*net.UDPConn, error) {
	addr := ":" + backupPort
	s, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("Error resolving UDP address: %v", err)
		return nil, err
	}
	return net.ListenUDP("udp", s)
}

// startPrimaryProcess initializes the primary process with UDP connection
func startPrimaryProcess(id, port, backupPort string) {
	backupAddr, err := net.ResolveUDPAddr("udp", ":"+backupPort)
	if err != nil {
		log.Printf("Failed to resolve backup address: %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, backupAddr)
	if err != nil {
		log.Printf("Failed to dial backup: %v", err)
		return
	}
	processPairs.PrimarySetup(id, port, backupPort, conn)
}

// spawnBackupProcess launches the backup process in a new terminal
func spawnBackupProcess(id, port string) {
	time.Sleep(1 * time.Second)
	err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
		"-id", id+"-backup", "-port", port).Run()
	if err != nil {
		log.Printf("Failed to spawn backup: %v", err)
	}
}

// monitorAndTakeOver handles backup taking over as primary when needed
func monitorAndTakeOver(id, port string, connection *net.UDPConn, backupPort string) {
	processPairs.BackupSetup(id, port, connection, backupPort)
	go func() {
		err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
			"-id", id, "-port", port).Run()
		if err != nil {
			log.Printf("Failed to spawn new backup: %v", err)
		}
	}()
	go startPrimaryProcess(id, port, backupPort)
}

// atoi safely converts string to int
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0 // Default to 0 if conversion fails
	}
	return i
}
