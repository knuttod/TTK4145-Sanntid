package main

import (
	"Heis/pkg/network/localip"
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/orders"
	"Heis/pkg/network/network"
	"flag"
	"fmt"
	"os"
	"log"
)

func main() {
	id, port := parseFlags()
	id = generateIDIfEmpty(id)

	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	elevio.Init("localhost:"+port, cfg.NumFloors)

	// Create channels for inter-module communication
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
	go orders.OrderHandler(id, localAssignedOrderCh, buttonPressCH, completedOrderCh,
		remoteElevatorCh, peerUpdateCh, fsmToOrdersCH, ordersToPeersCH)


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
