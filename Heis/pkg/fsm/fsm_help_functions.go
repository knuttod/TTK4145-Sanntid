package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
)

//Creates an elevator struct and ensures the elevator starts in a floor. Lights are cleared.
func fsmInit(drvFloorsCh chan int) elevator.Elevator {
	
	//initialize elevator struct
	elev := elevator.Elevator_init(numFloors, numBtns)

	//clear all lights
	for floor := range numFloors {
		for btn := range numBtns {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			elevio.SetDoorOpenLamp(false)
		}
	}

	//If elevator starts between two floors
	if elevio.GetFloor() == -1 {
		elevio.SetMotorDirection(elevio.MD_Down)

		//wait to arrive on a floor
		newFloor := <-drvFloorsCh
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevio.SetFloorIndicator(newFloor)
		elev.Floor = newFloor

	} else {
		floor := <- drvFloorsCh
		elevio.SetFloorIndicator(floor)
		elev.Floor = floor
	}

	return elev
}

// checks if there are any local orders on the floors above the elevator
func LocalOrderAbove(elev elevator.Elevator) bool {
	for floor := elev.Floor + 1; floor < numFloors; floor++ {
		for btn := range numBtns {
			if elev.LocalOrders[floor][btn] {
				return true
			}
		}
	}

	return false
}

// checks if there are any local orders on the floors below the elevator
func LocalOrderBelow(elev elevator.Elevator) bool {
	for floor := 0; floor < elev.Floor; floor++ {
		for btn := range numBtns {
			if elev.LocalOrders[floor][btn] {
				return true
			}
		}
	}
	return false
}

// checks if there are any local orders on the floor of the elevator
func localOrderHere(elev elevator.Elevator) bool {
	for btn := range numBtns{
		if elev.LocalOrders[elev.Floor][btn] {
			return true
		}
	}
	return false
}

// Chooses a direction for the elevator to move in according to the direction, behaviour and localOrders of the elevator
func ChooseDirection(elev elevator.Elevator) elevator.DirnBehaviourPair {
	switch elev.Dirn {
	case elevio.MD_Up:
		if LocalOrderAbove(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else if localOrderHere(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_DoorOpen}
		} else if LocalOrderBelow(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}
	case elevio.MD_Down:
		if LocalOrderBelow(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else if localOrderHere(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_DoorOpen}
		} else if LocalOrderAbove(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if localOrderHere(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_DoorOpen}
		} else if LocalOrderAbove(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else if LocalOrderBelow(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}
	default:
		return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
	}

}

// Checks if the elevator should stop on a floor or keep going
func ShouldStop(elev elevator.Elevator) bool {
	switch elev.Dirn {
	case elevio.MD_Down:
		if (elev.LocalOrders[elev.Floor][elevio.BT_HallDown]) || (elev.LocalOrders[elev.Floor][elevio.BT_Cab]) || (!LocalOrderBelow(elev)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Up:
		if (elev.LocalOrders[elev.Floor][elevio.BT_HallUp]) || (elev.LocalOrders[elev.Floor][elevio.BT_Cab]) || (!LocalOrderAbove(elev)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Stop:
		return true

	default:
		return true
	}
}

// Checks if a new order should be cleared immedeatly, e.g. a cab order in the same floor the elevator is on
func ShouldClearImmediately(elev elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		if elev.Floor == btn_floor {
			return true
		} else {
			return false
		}

	case elevator.CV_InDirn:
		if (elev.Floor == btn_floor) && ((elev.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(elev.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			(elev.Dirn == elevio.MD_Stop) || (btn_type == elevio.BT_Cab)) {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

// Clears at current floor and sends that the order is complete to the order module.
func ClearAtCurrentFloor(elev elevator.Elevator, completedOrderCH chan elevio.ButtonEvent) elevator.Elevator {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		for btn := range numBtns{
			elev = clearLocalOrder(elev, elev.Floor, elevio.ButtonType(btn), completedOrderCH)
		}

	case elevator.CV_InDirn:
		elev = clearLocalOrder(elev, elev.Floor, elevio.BT_Cab, completedOrderCH)
		switch elev.Dirn {
		case elevio.MD_Up:
			if (!LocalOrderAbove(elev)) && (!elev.LocalOrders[elev.Floor][elevio.BT_HallUp]) {
				elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallDown, completedOrderCH)
			}
			elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallUp, completedOrderCH)

		case elevio.MD_Down:
			if (!LocalOrderBelow(elev)) && (!elev.LocalOrders[elev.Floor][elevio.BT_HallDown]) {
				elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallUp, completedOrderCH)
			}
			elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallDown, completedOrderCH)

		case elevio.MD_Stop:
			elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallUp, completedOrderCH)
			elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallDown, completedOrderCH)
		}
	}
	return elev
}

// Sets the corresponding order in the localOrder 2D slice to true
func setLocalOrder(elev elevator.Elevator, floor int, btn elevio.ButtonType) elevator.Elevator {
	elev.LocalOrders[floor][btn] = true
	return elev
}

// Sets the corresponding order in the localOrder 2D slice to false and sends message of completed order to Orders
func clearLocalOrder(elev elevator.Elevator, floor int, btn elevio.ButtonType, completedOrderCH chan elevio.ButtonEvent) elevator.Elevator {
	elev.LocalOrders[floor][btn] = false
	//send on channel to orders that an order is completed/cleared
	completedOrderCH <- elevio.ButtonEvent{
		Floor:  floor,
		Button: btn,
	}
	return elev
}

// Clears all hall orders in localOrders
func removeLocalHallOrders(elev elevator.Elevator) elevator.Elevator {
	for floor := range numFloors {
		for btn := range numBtns - 1 {
			elev.LocalOrders[floor][btn] = false
		}
	}
	return elev
}

// Sets cab lights
func setCabLights(elev elevator.Elevator) {
	for floor := range numFloors{
		btn := int(elevio.BT_Cab)
		if elev.LocalOrders[floor][btn] {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
		} else {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
}


