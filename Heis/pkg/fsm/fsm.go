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
	numFloors          int
	numBtns         	 int
	doorTimerInterval time.Duration
	motorStopTimeout  time.Duration
)

// inits global variables from the config file
func init() {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	numFloors = cfg.NumFloors // Preserving your exact naming
	numBtns = cfg.NumBtns
	doorTimerInterval = cfg.DoorOpenDuration * time.Second
	motorStopTimeout = cfg.MotorStopTimeout * time.Second
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

	elev := fsmInit(id, drvFloorsCh)

	for {
		//sends a deepcopy to ensure correct message passing
		fsmToOrdersCH <- deepcopy.DeepCopyElevatorStruct(elev)
		select {
		//Inputs (buttons pressed) on each elevator is channeled to their respective local request
		case button_input := <-drvButtonsCh:
			buttonPressCH <- button_input

		//When an assigned order on a local elevator is channeled, it is set as an order to requestButtonPress that makes the elevators move
		case Order := <-localAssignedOrderCH:
			requestButtonPress(&elev, Order.Floor, Order.Button, doorTimerStartCh, departureFromFloorCh, completedOrderCH)

		case newFloor := <-drvFloorsCh:
			floorArrival(&elev, newFloor, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCH)

		case obstruction := <-drvObstrCh:
			if obstruction {
				fmt.Println("Obstruction switch activated")
				elev.Obstructed = true
				//remove hall orders since other elevators (this elevator if it is the only one on the nettwork) takes over from this
				elev = removeLocalHallOrders(elev)
			} else {
				elev.Obstructed = false
				doorTimerStartCh <- true
			}

		case <-motorStopCh:
			fmt.Println("motorstop")
			elev.MotorStop = true
			//remove hall orders since other elevators (this elevator if it is the only one on the nettwork) takes over from this
			elev = removeLocalHallOrders(elev)

		case <-doorTimerFinishedCh:
			if !elev.Obstructed && (elev.Behaviour != elevator.EB_Moving) {
				DoorTimeout(&elev, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCH)
			}
		}
	}
}
