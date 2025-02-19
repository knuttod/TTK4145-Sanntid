package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/network/bcast"
	"Heis/pkg/network/localip"
	"Heis/pkg/network/peers"
	"Heis/pkg/timer"
	"Heis/pkg/msgTypes"
	"Heis/pkg/elevator"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

//Public funksjoner har stor bokstav!!!!!!! Private har liten !!!!!
//!!!!!!!!!!!



func transmitState(e *elevator.Elevator, Tx chan msgTypes.UdpMsg, id string) {
	elevatorStateMsg := msgTypes.ElevatorStateMsg{
			Elevator: e,
			Id:       id,
		}
	for {
		Tx <- msgTypes.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
		time.Sleep(10 * time.Millisecond)
	}
}


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

	var e elevator.Elevator
	elevator.Elevator_init(&e, NumFloors, NumButtons)

	

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
	peerTxEnable := make(chan bool)
	Tx := make(chan msgTypes.UdpMsg)
	Rx := make(chan msgTypes.UdpMsg)	//Kanskje ha buffer her. For å få inn meldinger fra flere heiser samtidig. 

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Transmitter(16569, Tx)
	go bcast.Receiver(16569, Rx)

	go fsm.Fsm(&e, drv_buttons, drv_floors, drv_obstr, drv_stop, drv_doorTimerStart, drv_doorTimerFinished, Tx, Rx, peerTxEnable, peerUpdateCh, id)
	go timer.Timer(drv_doorTimerStart, drv_doorTimerFinished)

	go transmitState(&e, Tx, id)

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case a := <-Rx:
			fmt.Printf("Received: %#v\n", a)
		}
	}

	//	select {}

}
