package fsm

import (
	"Heis/pkg/config"
	"Heis/pkg/deepcopy"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
	"log"
	"time"
	// "fmt"
)

// define in config
var (
	N_floors          int
	N_buttons         int
	doorTimerInterval time.Duration
	motorStopTimeout  time.Duration
)

// inits global variables from the config file
func init() {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	N_floors = cfg.N_floors // Preserving your exact naming
	N_buttons = cfg.N_buttons
	doorTimerInterval = cfg.DoorOpenDuration * time.Second
	motorStopTimeout = cfg.MotorStopTimeout * time.Millisecond
}

// FSM handles core logic of a single Elevator. Interacts with orders via localAssignedOrderCH, localRequestCH and completedOrderCH.
// Also takes input from elevio on drv channels. Interacts with external timer on doorTimerStartCH and doorTimerFinishedCH
func Fsm(id string, localAssignedOrderCH, buttonPressCH, completedOrderCH chan elevio.ButtonEvent, fsmToOrdersCH chan elevator.Elevator) {

	drvButtonsCh := make(chan elevio.ButtonEvent)
	drvFloorsCh := make(chan int)
	drvObstrCh := make(chan bool)
	drvStopCh := make(chan bool)

	doorTimerStartCh := make(chan bool)
	doorTimerFinishedCh := make(chan bool)

	arrivedOnFloorCh := make(chan bool)
	departureFromFloorCh := make(chan bool)
	motorStopCh := make(chan bool)

	go elevio.PollButtons(drvButtonsCh)
	go elevio.PollFloorSensor(drvFloorsCh)
	go elevio.PollObstructionSwitch(drvObstrCh)
	go elevio.PollStopButton(drvStopCh) //kanskje implementere stop?

	go doorTimer(doorTimerInterval, doorTimerStartCh, doorTimerFinishedCh)
	go motorStopTimer(motorStopTimeout, arrivedOnFloorCh, departureFromFloorCh, motorStopCh)

	elev := elevator.Elevator_init(N_floors, N_buttons, id)

	for floor := range N_floors {
		for btn := range N_buttons {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			elevio.SetDoorOpenLamp(false)
		}
	}

	floor := elevio.GetFloor()
	if floor == -1 {
		initBetweenFloors(&elev)
		current_floor := <-drvFloorsCh
		elevio.SetMotorDirection(elevio.MD_Stop)
		elev.Floor = current_floor
		elevio.SetFloorIndicator(current_floor)
		elev.Dirn = elevio.MD_Stop
		elev.Behaviour = elevator.EB_Idle

		// floorArrival(&elev, current_floor, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCH)
	} else {
		elev.Floor = floor
		elevio.SetFloorIndicator(floor)
	}

	// //trenger kanskje ikke denne?
	// fsmToOrdersCH <- deepcopy.DeepCopyElevatorStruct(elev)

	for {
		fsmToOrdersCH <- deepcopy.DeepCopyElevatorStruct(elev)
		select {
		//Inputs (buttons pressed) on each elevator is channeled to their respective local request
		case button_input := <-drvButtonsCh:
			// fmt.Println("btn")
			buttonPressCH <- button_input

		//When an assigned order on a local elevator is channeled, it is set as an order to requestButtonPress that makes the elevators move
		case Order := <-localAssignedOrderCH:
			requestButtonPress(&elev, Order.Floor, Order.Button, doorTimerStartCh, departureFromFloorCh, completedOrderCH)
		case current_floor := <-drvFloorsCh:
			floorArrival(&elev, current_floor, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCH)
		case obstruction := <-drvObstrCh:
			if obstruction {
				elev.Obstructed = true
				//remove hall orders since other elevators (this elevator if it is the only one on the nettwork) takes over from this
				elev = removeLocalHallOrders(elev)
				fmt.Println("Obstruction switch activated")
			} else {
				elev.Obstructed = false
				doorTimerStartCh <- true
			}
		case <-motorStopCh:
			fmt.Println("motorstop")
			elev.MotorStop = true
			elev = removeLocalHallOrders(elev)

		case <-doorTimerFinishedCh:
			if !elev.Obstructed && (elev.Behaviour != elevator.EB_Moving) {
				DoorTimeout(&elev, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCH)
			}
		}
	}
}
