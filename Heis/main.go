package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/network/localip"
	"Heis/pkg/network/network"
	"Heis/pkg/orders"
	"Heis/pkg/processPairs"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	id, port, processPairsFlag := parseFlags()
	id = generateIDIfEmpty(id)
	backupPort := calculateBackupPort(port)

	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Check if user have selected the processPair option
	if processPairsFlag {
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
	} else {

		elevio.Init("localhost:"+port, cfg.NumFloors)

		peerUpdateCh := make(chan network.PeerUpdate)
		remoteElevatorCh := make(chan network.ElevatorStateMsg)
		peerTxEnable := make(chan bool)
		localAssignedOrderCh := make(chan elevio.ButtonEvent)
		buttonPressCH := make(chan elevio.ButtonEvent)
		completedOrderCh := make(chan elevio.ButtonEvent)
		fsmToOrdersCH := make(chan elevator.Elevator)
		ordersToPeersCH := make(chan elevator.NetworkElevator)
		transmitterToRecivierSkipCh := make(chan bool)

		// Launch main elevator system components as goroutines
		go network.Transmitter(17135, id, peerTxEnable, transmitterToRecivierSkipCh, ordersToPeersCH)
		go network.Receiver(17135, id, transmitterToRecivierSkipCh, peerUpdateCh, remoteElevatorCh)
		go fsm.Fsm(id, localAssignedOrderCh, buttonPressCH, completedOrderCh, fsmToOrdersCH)
		go orders.OrderHandler(id, localAssignedOrderCh, buttonPressCH, completedOrderCh, remoteElevatorCh, peerUpdateCh, fsmToOrdersCH, ordersToPeersCH)

	}

	select {}
}

// parseFlags handles command-line flag parsing
func parseFlags() (string, string, bool) {
	var id, port string
	var processPairsFlag bool
	flag.BoolVar(&processPairsFlag, "processPairsFlag", false, "use the original main function instead of the new one")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()
	return id, port, processPairsFlag
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
