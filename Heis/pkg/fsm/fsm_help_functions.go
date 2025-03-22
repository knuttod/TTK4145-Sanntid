package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
	// "Heis/pkg/timer"
	//"fmt"
)

// When initialising between floors, the motor direction is set downwards.
// When it reaches a floor it will then behave normaly.
func initBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = elevator.EB_Moving
}

// Handles button presses on a local level, by processing requests based on the
// elevator's current behavior. If the elevator is idle, it determines the next action
// (moving or opening doors). If the elevator is moving or has doors open, it updates
// the request state accordingly. The function also manages the door timer, sends updated
// elevator states over UDP, and updates the button lights.
func requestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64, floorArrivalCh, motorTimoutStartCh chan bool, completedOrderCH chan elevio.ButtonEvent) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			// send clear to assigned orders
			completedOrderCH <- elevio.ButtonEvent{
				Floor:  btn_floor,
				Button: btn_type,
			}
		} else {
			(*e).LocalOrders[btn_floor][btn_type] = true
		}

	case elevator.EB_Moving:
		(*e).LocalOrders[btn_floor][btn_type] = true

	case elevator.EB_Idle:
		(*e).LocalOrders[btn_floor][btn_type] = true
		var pair elevator.DirnBehaviourPair = ChooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			(*e).LocalOrders = ClearAtCurrentFloor((*e), completedOrderCH).LocalOrders

		case elevator.EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
			//clear something at this floor??

			motorTimoutStartCh <- true

		case elevator.EB_Idle:
			//need something here?
		}

	}
	setCabLights(e)
}

// When arriving at a floor this sets the floor indicator to the floor, and checks if it is supposed
// to stop. if it is supposed to stop it stops, clears the floor then opens the door.
func floorArrival(e *elevator.Elevator, newFloor int, drv_doorTimer chan float64, floorArrivalCh, motorTimoutStartCh chan bool, completedOrderCH chan elevio.ButtonEvent) {

	(*e).Floor = newFloor
	elevio.SetFloorIndicator((*e).Floor)

	if !(*e).MotorStop {
		select {
		case floorArrivalCh <- true:
		default:
		}

	}
	if (*e).MotorStop {
		fmt.Println("power back")
		(*e).MotorStop = false
	}

	switch (*e).Behaviour {
	case elevator.EB_Moving:
		if ShouldStop((*e)) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			(*e).LocalOrders = ClearAtCurrentFloor((*e), completedOrderCH).LocalOrders
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			setCabLights(e)
			(*e).Behaviour = elevator.EB_DoorOpen
		} else {
			motorTimoutStartCh <- true

		}
	}
}

// DoorTimeout is used to start the timer when the door openes, and
// handles if there is an obstruction when the door closes. It is
// runned twice, once at the begining of the timer initialisation and
// once when the door is supposed to close to check if the obstruction
// is active.
func DoorTimeout(e *elevator.Elevator, drv_doorTimer chan float64, floorArrivalCh, motorTimoutStartCh chan bool, completedOrderCH chan elevio.ButtonEvent) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		var pair elevator.DirnBehaviourPair = ChooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch (*e).Behaviour {
		case elevator.EB_DoorOpen:
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s //????
			//drv_doorTimer <- 0.0
			(*e).LocalOrders = ClearAtCurrentFloor((*e), completedOrderCH).LocalOrders
			setCabLights(e)

		//lagt inn selv
		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)
			motorTimoutStartCh <- true
		//

		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)

		}
	}
}

// Updates the elevator's button lights to reflect the current request states.
// Turns on lights for active LocalOrders and turns them off otherwise.
func setCabLights(e *elevator.Elevator) {
	//set ligths
	for floor := 0; floor < N_floors; floor++ {
		// for btn := 0; btn < N_buttons; btn++ {
		btn := int(elevio.BT_Cab)
		if e.LocalOrders[floor][btn] {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
		} else {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
		// }
	}
}

func clearLocalHallOrders(e elevator.Elevator) elevator.Elevator {
	for floor := range N_floors {
		for btn := range (N_buttons-1) {
			e.LocalOrders[floor][btn] = false
		}
	}
	return e
}
