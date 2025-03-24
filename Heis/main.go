package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/network/msgTypes"
	"Heis/pkg/network/localip"
	"Heis/pkg/network/peers"
	"Heis/pkg/orders"

	// "Heis/pkg/message"
	"flag"
	"fmt"
	"log"
	"os"
)



func main() {
	// Elevator id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`. Port should match that of elevatorserver
	var id string
	var port string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// Use the loaded configuration
	NumFloors := cfg.NumFloors
	// NumButtons := cfg.NumButtons
	//temp
	// NumElevators := 3


	
	elevio.Init("localhost:"+port, NumFloors)


	peerUpdateCh := make(chan peers.PeerUpdate)
	remoteElevatorCh := make(chan msgTypes.ElevatorStateMsg)
	peerTxEnable := make(chan bool)

	localAssignedOrderCH := make(chan elevio.ButtonEvent) //might need buffer
	buttonPressCH := make(chan elevio.ButtonEvent) //might need buffer
	completedOrderCh := make(chan elevio.ButtonEvent, 1) //can clear up to 3 orders at a time

	fsmToOrdersCH := make(chan elevator.Elevator)
	ordersToPeersCH := make(chan elevator.NetworkElevator)
	nettworkDisconnectCh := make(chan bool)

	transmitterToRecivierSkipCh := make(chan bool)


	go peers.Transmitter(17135, id, peerTxEnable, transmitterToRecivierSkipCh, ordersToPeersCH)
	go peers.Receiver(17135, id, transmitterToRecivierSkipCh, peerUpdateCh, remoteElevatorCh)
	
	go fsm.Fsm(id, localAssignedOrderCH, buttonPressCH, completedOrderCh, fsmToOrdersCH)
	
	go orders.OrderHandler(id, localAssignedOrderCH, buttonPressCH, completedOrderCh, remoteElevatorCh, peerUpdateCh, nettworkDisconnectCh,fsmToOrdersCH, ordersToPeersCH)

	select{}
}
