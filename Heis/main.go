package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/orders"
	"Heis/pkg/network/localip"
	"Heis/pkg/network/network"
	"Heis/pkg/processpairs"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {

	// Load arguments from command flags
	id, port, processPairsFlag := parseFlags()
	id = generateIDIfEmpty(id)
	
	// Load parameters from config file
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Check if user have selected the processPair option.
	// The processPair option will only work on linux systems.
	if processPairsFlag {
		backupPort := calculateBackupPort(port)
		// Determine process role using processpairs
		connection, err := processpairs.SetupUDPListener(backupPort)
		// If it can not connect to a primary it should take over as primary
		if err != nil {
			fmt.Println("Starting as primary...")
			go processpairs.StartPrimaryProcess(id, port, backupPort)
			go processpairs.SpawnBackupProcess(id, port)
		} else {
			fmt.Println("Starting as backup...")
			go processpairs.BackupSetup(id, port, connection, backupPort)
			go processpairs.MonitorAndTakeOver(id, port, connection, backupPort)
		}

	// Otherwise the system will run without processpairs.
	} else {
		elevio.Init("localhost:"+port, cfg.NumFloors)

		// between fsm and orders
		buttonPressCH := make(chan elevio.ButtonEvent)
		completedOrderCh := make(chan elevio.ButtonEvent)
		fsmToOrdersCH := make(chan elevator.Elevator)
		localAssignedOrderCh := make(chan elevio.ButtonEvent)

		// between orders and network
		ordersToPeersCH := make(chan elevator.NetworkElevator)
		peerUpdateCh := make(chan network.PeerUpdate)
		remoteElevatorUpdateCh := make(chan network.ElevatorStateMsg)

		// enable sending on network
		peerTxEnable := make(chan bool)
		// between transmitter and reciever
		transmitterToRecivierSkipCh := make(chan bool)

		// Launch main elevator system components as goroutines
		go fsm.Fsm(id, localAssignedOrderCh, buttonPressCH, completedOrderCh, fsmToOrdersCH)
		go orders.OrderHandler(id, localAssignedOrderCh, buttonPressCH, completedOrderCh, remoteElevatorUpdateCh, peerUpdateCh, fsmToOrdersCH, ordersToPeersCH)
		go network.Transmitter(17135, id, peerTxEnable, transmitterToRecivierSkipCh, ordersToPeersCH)
		go network.Receiver(17135, id, transmitterToRecivierSkipCh, peerUpdateCh, remoteElevatorUpdateCh)

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
		// Default to 0 if conversion fails
		return 0 
	}
	return i
}
