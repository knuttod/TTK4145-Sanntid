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
	DoorTimerInterval time.Duration
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
	DoorTimerInterval = cfg.DoorOpenDuration * time.Second
	motorStopTimeout = cfg.MotorStopTimeout * time.Second
}

// FSM handles core logic of a single Elevator. 
// Interacts with orders via localAssignedOrderCh, localRequestCH and completedOrderCh.
// Also takes input from elevio on drv channels. 
func Fsm(id string, localAssignedOrderCh, buttonPressCh, completedOrderCh chan elevio.ButtonEvent, fsmToOrdersCh chan elevator.Elevator) {

	// Elevio
	drvButtonsCh := make(chan elevio.ButtonEvent)
	drvFloorsCh := make(chan int)
	drvObstrCh := make(chan bool)
	drvStopCh := make(chan bool)

	// Door timer
	doorTimerStartCh := make(chan bool)
	doorTimerFinishedCh := make(chan bool)

	// Motor stop timer
	arrivedOnFloorCh := make(chan bool)
	departureFromFloorCh := make(chan bool)
	motorStopCh := make(chan bool)

	go elevio.PollButtons(drvButtonsCh)
	go elevio.PollFloorSensor(drvFloorsCh)
	go elevio.PollObstructionSwitch(drvObstrCh)
	go elevio.PollStopButton(drvStopCh)

	go doorTimer(DoorTimerInterval, doorTimerStartCh, doorTimerFinishedCh)
	go motorStopTimer(motorStopTimeout, arrivedOnFloorCh, departureFromFloorCh, motorStopCh)

	elev := fsmInit(id, drvFloorsCh)

	for {
		//sends a deepcopy to ensure correct message passing
		fsmToOrdersCh <- deepcopy.DeepCopyElevatorStruct(elev)
		select {
		//Inputs (buttons pressed) on each elevator is channeled to their respective local request
		case button_input := <-drvButtonsCh:
			buttonPressCh <- button_input

		//When an assigned order on a local elevator is channeled, it is set as an order to requestButtonPress that makes the elevators move
		case Order := <-localAssignedOrderCh:
			elev = requestButtonPress(elev, Order.Floor, Order.Button, doorTimerStartCh, departureFromFloorCh, completedOrderCh)

		case newFloor := <-drvFloorsCh:
			elev = floorArrival(elev, newFloor, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCh)

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
				elev = doorTimeout(elev, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh, completedOrderCh)
			}
		}
	}
}


// Handles button presses on a local level, by processing requests based on the
// elevator's current behavior. If the elevator is idle, it determines the next action
// (moving or opening doors). If the elevator is moving or has doors open, it updates
// the request state accordingly. The function also manages the door timer, sends updated
// elevator states over UDP, and updates the button lights.
func requestButtonPress(elev elevator.Elevator, btnFloor int, btnType elevio.ButtonType, 
	doorTimerStartCh, departureFromFloorCh chan bool, completedOrderCh chan elevio.ButtonEvent) elevator.Elevator {

	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		if ShouldClearImmediately(elev, btnFloor, btnType) {
			doorTimerStartCh <- true
			elev = clearLocalOrder(elev, btnFloor, btnType, completedOrderCh)
		} else {
			elev = setLocalOrder(elev, btnFloor, btnType)
		}

	case elevator.EB_Moving:
		elev = setLocalOrder(elev, btnFloor, btnType)

	case elevator.EB_Idle:
		elev = setLocalOrder(elev, btnFloor, btnType)

		var directionAndBehaviour elevator.DirnBehaviourPair = ChooseDirection(elev)
		elev.Dirn = directionAndBehaviour.Dirn
		elev.Behaviour = directionAndBehaviour.Behaviour

		switch directionAndBehaviour.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			doorTimerStartCh <- true
			elev.LocalOrders = ClearAtCurrentFloor(elev, completedOrderCh).LocalOrders

			// To make sure both hall call up and down are not cleared when an elevator has no orders and gets both calls in the floor it is currently at
			if btnType == elevio.BT_HallUp {
				elev.Dirn = elevio.MD_Up
			} else if btnType == elevio.BT_HallDown {
				elev.Dirn = elevio.MD_Down
			}

		case elevator.EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)
			departureFromFloorCh <- true

		case elevator.EB_Idle:
			//nothing should be done
		}
	}
	setCabLights(elev)
	return elev
}

// When arriving at a floor this sets the floor indicator to the floor, and checks if it is supposed
// to stop. if it is supposed to stop it stops, clears the floor then opens the door.
func floorArrival(elev elevator.Elevator, newFloor int, 
	doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh chan bool, completedOrderCh chan elevio.ButtonEvent) elevator.Elevator {

	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	if !elev.MotorStop {
		arrivedOnFloorCh <- true
	} else if elev.MotorStop {
		fmt.Println("power back")
		elev.MotorStop = false
	}

	switch elev.Behaviour {
	case elevator.EB_Moving:
		if ShouldStop(elev) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			doorTimerStartCh <- true
			elev.LocalOrders = ClearAtCurrentFloor(elev, completedOrderCh).LocalOrders
			setCabLights(elev)
			elev.Behaviour = elevator.EB_DoorOpen
		} else {
			departureFromFloorCh <- true
		}
	}
	return elev
}

// DoorTimeout is used to start the timer when the door openes, and
// handles if there is an obstruction when the door closes. It is
// runned twice, once at the begining of the timer initialisation and
// once when the door is supposed to close to check if the obstruction
// is active.
func doorTimeout(elev elevator.Elevator, doorTimerStartCh, arrivedOnFloorCh, departureFromFloorCh chan bool, completedOrderCh chan elevio.ButtonEvent) elevator.Elevator {

	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		var directionAndBehaviour elevator.DirnBehaviourPair = ChooseDirection(elev)
		elev.Dirn = directionAndBehaviour.Dirn
		elev.Behaviour = directionAndBehaviour.Behaviour

		switch elev.Behaviour {
		case elevator.EB_DoorOpen:
			doorTimerStartCh <- true
			elev.LocalOrders = ClearAtCurrentFloor(elev, completedOrderCh).LocalOrders
			setCabLights(elev)

		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
			departureFromFloorCh <- true

		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)

		}
	}
	return elev
}

