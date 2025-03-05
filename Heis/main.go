package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/bcast"
	"Heis/pkg/network/localip"
	"Heis/pkg/network/peers"
	"Heis/pkg/orders"
	"Heis/pkg/timer"
	"Heis/pkg/message"
	"flag"
	"fmt"
	"log"
	"os"
)

//Public funksjoner har stor bokstav!!!!!!! Private har liten !!!!!
//!!!!!!!!!!!

// Kanskje ha egne funksjoner som spawner andre go routines?? Typ Gruppere sammen de funksjonene som "jobber sammen" i en funksjon of sette opp channels of funksjoner i dem
// Typ ha et tre med go funksjoner og grener som channels som hører sammen


func main() {
	// Elevator id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	var port string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
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
	NumButtons := cfg.NumButtons
	// NumFloors := 4

	NumElevators := 3

	var e elevator.Elevator
	elevator.Elevator_init(&e, NumFloors, NumButtons, NumElevators, id)

	

	elevio.Init("localhost:"+port, NumFloors)


	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_doorTimerStart := make(chan float64)
	drv_doorTimerFinished := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	peerUpdateCh := make(chan peers.PeerUpdate)
	elevatorStateCh := make(chan msgTypes.ElevatorStateMsg)
	peerTxEnable := make(chan bool)

	Tx := make(chan msgTypes.UdpMsg)
	Rx := make(chan msgTypes.UdpMsg)

	go peers.Transmitter(15647, id, peerTxEnable, &e)
	go peers.Receiver(15647, peerUpdateCh, elevatorStateCh)
	go bcast.Transmitter(16569, Tx)
	go bcast.Receiver(16569, Rx)

	remoteElevators := make(map[string]elevator.Elevator)

	go timer.Timer(drv_doorTimerStart, drv_doorTimerFinished)

	
	go message.TransmitState(&e, Tx, id)
	// go orders.LocalButtonPressHandler(&e, drv_buttons, LocalOrderRequest)
	// go orders.GlobalOrderMerger(&e, stateRx, stateTx, LocalOrderRequest, LocalOrderTask)
	// go orders.SyncGlobalWithLocalOrders(&e)
	// go orders.CheckForCompletedOrders(&e, LocalOrderRequest)

	localAssignedOrder := make(chan elevio.ButtonEvent)
	localRequest := make(chan elevio.ButtonEvent)

	completedOrderCh := make(chan elevio.ButtonEvent)

	go fsm.Fsm(&e, drv_buttons, drv_floors, drv_obstr, drv_stop, drv_doorTimerStart, drv_doorTimerFinished, id, localAssignedOrder, localRequest, completedOrderCh)
	go orders.OrderHandler(&e, &remoteElevators, localAssignedOrder, localRequest, elevatorStateCh, completedOrderCh)


	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		// case a := <-Rx:
		// 	fmt.Printf("Received: %#v\n", a)
		}
	}

}
