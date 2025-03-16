package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/deepcopy"
	"fmt"
	// "fmt"
)

// define in config
const N_floors = 4
const N_buttons = 3


// FSM handles core logic of a single Elevator. Interacts with orders via localAssignedOrderCH, localRequestCH and completedOrderCH. 
// Also takes input from elevio on drv channels. Interacts with external timer on doorTimerStartCH and doorTimerFinishedCH
func Fsm(elev *elevator.Elevator, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr,
	drv_stop chan bool, doorTimerStartCH chan float64, doorTimerFinishedCH chan bool,
	id string, localAssignedOrderCH chan elevio.ButtonEvent, buttonPressCH, completedOrderCH chan elevio.ButtonEvent, fsmToOrdersCH chan elevator.Elevator) {

	floor := elevio.GetFloor()
	if floor == -1 {
		initBetweenFloors(elev)
		for (*elev).Floor == -1 {
			current_floor := <-drv_floors
			floorArrival(elev, current_floor, doorTimerStartCH, completedOrderCH)
		} 
	} else {
		(*elev).Floor = floor
	}

	fmt.Println("startFloor: ",(*elev).Floor)

	fsmToOrdersCH <- *elev

	for {
		select {
		// case fsmToOrdersCH <- *elev:
		case fsmToOrdersCH <- deepcopy.DeepCopyElevatorStruct(*elev):
		default:
		}

		select {
		//Inputs (buttons pressed) on each elevator is channeled to their respective local request
		case button_input := <-drv_buttons:
			buttonPressCH <- button_input

		//When an assigned order on a local elevator is channeled, it is set as an order to requestButtonPress that makes the elevators move
		case Order := <-localAssignedOrderCH:
			requestButtonPress(elev, Order.Floor, Order.Button, doorTimerStartCH, completedOrderCH)

		case current_floor := <-drv_floors:
			floorArrival(elev, current_floor, doorTimerStartCH, completedOrderCH)

			// fmt.Printf("drv_floors: %v", current_floor)
		case obstruction := <-drv_obstr:
			if obstruction {
				(*elev).Obstructed = true
				(*elev).Behaviour = elevator.EB_Unavailable
			} else {
				(*elev).Obstructed = false
				(*elev).Behaviour = elevator.EB_DoorOpen
				doorTimerStartCH <- (*elev).Config.DoorOpenDuration_s
			}

		case <-doorTimerFinishedCH:
			if !elev.Obstructed {
				DoorTimeout(elev, doorTimerStartCH, completedOrderCH)
				// fmt.Println("drv_doortimer timed out")
				DoorTimeout(elev, doorTimerStartCH, completedOrderCH)
			}
		default:
			//non blocking
		}

		

	}
}
