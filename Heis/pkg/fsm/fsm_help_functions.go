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
func initBetweenFloors(elev *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*elev).Dirn = elevio.MD_Down
	(*elev).Behaviour = elevator.EB_Moving
}

// Handles button presses on a local level, by processing requests based on the
// elevator's current behavior. If the elevator is idle, it determines the next action
// (moving or opening doors). If the elevator is moving or has doors open, it updates
// the request state accordingly. The function also manages the door timer, sends updated
// elevator states over UDP, and updates the button lights.
func requestButtonPress(elev *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, doorTimerStartCh chan bool, arrivedOnFloorCh chan bool, completedOrderCH chan elevio.ButtonEvent) {
	switch (*elev).Behaviour {
	case elevator.EB_DoorOpen:
		if ShouldClearImmediately((*elev), btn_floor, btn_type) {
			// send clear to assigned orders
			*elev = clearLocalOrder(*elev, btn_floor, btn_type, completedOrderCH)
			doorTimerStartCh <-true
		} else {
			*elev = setLocalOrder(*elev, btn_floor, btn_type)
			// (*elev).LocalOrders[btn_floor][btn_type] = true
		}

	case elevator.EB_Moving:
		*elev = setLocalOrder(*elev, btn_floor, btn_type)
		// (*elev).LocalOrders[btn_floor][btn_type] = true

	case elevator.EB_Idle:
		*elev = setLocalOrder(*elev, btn_floor, btn_type)
		// (*elev).LocalOrders[btn_floor][btn_type] = true
		var pair elevator.DirnBehaviourPair = ChooseDirection((*elev))
		(*elev).Dirn = pair.Dirn
		(*elev).Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			(*elev).LocalOrders = ClearAtCurrentFloor((*elev), completedOrderCH).LocalOrders
			elevio.SetDoorOpenLamp(true)
			doorTimerStartCh <- true

			// To make sure both hall call up and down are not cleared when an elevator has no orders and gets both calls in the floor it is currently at
			if btn_type == elevio.BT_HallUp {
				(*elev).Dirn = elevio.MD_Up
			} else if btn_type == elevio.BT_HallDown {
				(*elev).Dirn = elevio.MD_Down
			}

		case elevator.EB_Moving:
			elevio.SetMotorDirection((*elev).Dirn)
			arrivedOnFloorCh <- true

		case elevator.EB_Idle:
			//nothing should be done
		}

	}
	setCabLights(elev)
}

// When arriving at a floor this sets the floor indicator to the floor, and checks if it is supposed
// to stop. if it is supposed to stop it stops, clears the floor then opens the door.
func floorArrival(elev *elevator.Elevator, newFloor int, doorTimerStartCh chan bool, floorArrivalCh, arrivedOnFloorCh chan bool, completedOrderCH chan elevio.ButtonEvent) {

	(*elev).Floor = newFloor
	elevio.SetFloorIndicator((*elev).Floor)

	if !(*elev).MotorStop {
		select {
		case floorArrivalCh <- true:
		default:
		}

	}
	if (*elev).MotorStop {
		fmt.Println("power back")
		(*elev).MotorStop = false
	}

	switch (*elev).Behaviour {
	case elevator.EB_Moving:
		if ShouldStop((*elev)) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			(*elev).LocalOrders = ClearAtCurrentFloor((*elev), completedOrderCH).LocalOrders
			doorTimerStartCh <- true
			setCabLights(elev)
			(*elev).Behaviour = elevator.EB_DoorOpen
		} else {
			arrivedOnFloorCh <- true

		}
	}
}

// DoorTimeout is used to start the timer when the door openes, and
// handles if there is an obstruction when the door closes. It is
// runned twice, once at the begining of the timer initialisation and
// once when the door is supposed to close to check if the obstruction
// is active.
func DoorTimeout(elev *elevator.Elevator, doorTimerStartCh chan bool, floorArrivalCh, arrivedOnFloorCh chan bool, completedOrderCH chan elevio.ButtonEvent) {

	switch (*elev).Behaviour {
	case elevator.EB_DoorOpen:
		var pair elevator.DirnBehaviourPair = ChooseDirection((*elev))
		(*elev).Dirn = pair.Dirn
		(*elev).Behaviour = pair.Behaviour

		switch (*elev).Behaviour {
		case elevator.EB_DoorOpen:
			doorTimerStartCh <- true //????
			(*elev).LocalOrders = ClearAtCurrentFloor((*elev), completedOrderCH).LocalOrders
			setCabLights(elev)

		//lagt inn selv
		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*elev).Dirn)
			arrivedOnFloorCh <- true
		//

		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*elev).Dirn)

		}
	}
}

func setCabLights(elev *elevator.Elevator) {
	for floor := 0; floor < N_floors; floor++ {
		btn := int(elevio.BT_Cab)
		if elev.LocalOrders[floor][btn] {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
		} else {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
}

func removeLocalHallOrders(elev elevator.Elevator) elevator.Elevator {
	for floor := range N_floors {
		for btn := range (N_buttons-1) {
			elev.LocalOrders[floor][btn] = false
		}
	}
	return elev
}
