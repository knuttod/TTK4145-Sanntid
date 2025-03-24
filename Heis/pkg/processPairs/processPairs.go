package processPairs

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"
	"Heis/pkg/orders"
	"log"
	"net"
	"time"
)

// PrimarySetup configures and runs the elevator system as the primary process
func PrimarySetup(id string, port string, backupPort string, conn *net.UDPConn) {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	elevio.Init("localhost:"+port, cfg.NumFloors)

	// Create channels for inter-module communication
	peerUpdateCh := make(chan peers.PeerUpdate)
	remoteElevatorCh := make(chan msgTypes.ElevatorStateMsg)
	peerTxEnable := make(chan bool)
	localAssignedOrderCH := make(chan elevio.ButtonEvent)
	buttonPressCH := make(chan elevio.ButtonEvent)
	completedOrderCh := make(chan elevio.ButtonEvent, 3)
	fsmToOrdersCH := make(chan elevator.Elevator)
	ordersToPeersCH := make(chan elevator.NetworkElevator)
	networkDisconnectCh := make(chan bool)
	transmitterToRecivierSkipCh := make(chan bool)

	// Launch main elevator system components as goroutines
	go peers.Transmitter(17135, id, peerTxEnable, transmitterToRecivierSkipCh, ordersToPeersCH)
	go peers.Receiver(17135, id, transmitterToRecivierSkipCh, peerUpdateCh, remoteElevatorCh)
	go fsm.Fsm(id, localAssignedOrderCH, buttonPressCH, completedOrderCh, fsmToOrdersCH)
	go orders.OrderHandler(id, localAssignedOrderCH, buttonPressCH, completedOrderCh,
		remoteElevatorCh, peerUpdateCh, networkDisconnectCh, fsmToOrdersCH, ordersToPeersCH)

	// Periodic state sync to backup
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			_, err := conn.Write([]byte("ping")) // Simple heartbeat
			if err != nil {
				log.Printf("Failed to send to backup: %v", err)
			}
		}
	}()
}

// BackupSetup configures and runs the elevator system as the backup process
func BackupSetup(id string, port string, connection *net.UDPConn, backupPort string) {
	defer connection.Close()

	buffer := make([]byte, 1024)

	for {
		connection.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := connection.ReadFromUDP(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Primary failed, taking over as primary...")
				return
			}
		}
	}
}
