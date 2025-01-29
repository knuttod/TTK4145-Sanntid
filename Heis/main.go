package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/timer"
	"log"
)

//Public funksjoner har stor bokstav!!!!!!! Private har liten !!!!!
//!!!!!!!!!!!

func main() {

	cfg, err := config.LoadConfig("Heis/config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// Use the loaded configuration
	NumFloors := cfg.NumFloors
	NumButtons := cfg.NumButtons

	elevio.Init("localhost:15657", NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_doorTimer := make(chan float64)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go fsm.Fsm(drv_buttons, drv_floors, drv_obstr, drv_stop, drv_doorTimer)
	go timer.Timer(drv_doorTimer)

	//add functionality to resolve starting between floors

	//Initializing floor matrix
	floorMatrix := make([][]int, NumFloors)
	for i := range floorMatrix {
		floorMatrix[i] = make([]int, NumButtons)
	}

	select {}

}
