package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
)


func fsmInit(id string, drvFloorsCh chan int) elevator.Elevator {
	
	//initialize elevator struct
	elev := elevator.Elevator_init(numFloors, numBtns, id)

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


func Above(elev elevator.Elevator) bool {
	for f := elev.Floor + 1; f < numFloors; f++ {
		for btn := 0; btn < numBtns; btn++ {
			if elev.LocalOrders[f][btn] {
				return true
			}
		}
	}

	return false
}

func Below(elev elevator.Elevator) bool {
	for f := 0; f < elev.Floor; f++ {
		for btn := 0; btn < numBtns; btn++ {
			if elev.LocalOrders[f][btn] {
				return true
			}
		}
	}
	return false
}

func Here(elev elevator.Elevator) bool {
	for btn := 0; btn < numBtns; btn++ {
		if elev.LocalOrders[elev.Floor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirection(elev elevator.Elevator) elevator.DirnBehaviourPair {
	switch elev.Dirn {
	case elevio.MD_Up:
		if Above(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else if Here(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_DoorOpen}
		} else if Below(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}
	case elevio.MD_Down:
		if Below(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else if Here(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_DoorOpen}
		} else if Above(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if Here(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_DoorOpen}
		} else if Above(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else if Below(elev) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}
	default:
		return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
	}

}

func ShouldStop(elev elevator.Elevator) bool {
	//Får av en eller annen rar grunn etasje 4??
	switch elev.Dirn {
	case elevio.MD_Down:
		if (elev.LocalOrders[elev.Floor][elevio.BT_HallDown]) || (elev.LocalOrders[elev.Floor][elevio.BT_Cab]) || (!Below(elev)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Up:
		if (elev.LocalOrders[elev.Floor][elevio.BT_HallUp]) || (elev.LocalOrders[elev.Floor][elevio.BT_Cab]) || (!Above(elev)) {
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
		for btn := 0; btn < numBtns; btn++ {
			elev = clearLocalOrder(elev, elev.Floor, elevio.ButtonType(btn), completedOrderCH)
		}

	case elevator.CV_InDirn:
		elev = clearLocalOrder(elev, elev.Floor, elevio.BT_Cab, completedOrderCH)
		switch elev.Dirn {
		case elevio.MD_Up:
			if (!Above(elev)) && !(elev.LocalOrders[elev.Floor][elevio.BT_HallUp]) {
				elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallDown, completedOrderCH)
			}
			elev = clearLocalOrder(elev, elev.Floor, elevio.BT_HallUp, completedOrderCH)

		case elevio.MD_Down:
			if (!Below(elev)) && !(elev.LocalOrders[elev.Floor][elevio.BT_HallDown]) {
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

func setLocalOrder(elev elevator.Elevator, floor int, btn elevio.ButtonType) elevator.Elevator {
	elev.LocalOrders[floor][btn] = true
	return elev
}

func clearLocalOrder(elev elevator.Elevator, floor int, btn elevio.ButtonType, completedOrderCH chan elevio.ButtonEvent) elevator.Elevator {
	elev.LocalOrders[floor][btn] = false
	//send on channel to orders that an order is completed/cleared
	completedOrderCH <- elevio.ButtonEvent{
		Floor:  floor,
		Button: btn,
	}
	return elev
}

func removeLocalHallOrders(elev elevator.Elevator) elevator.Elevator {
	for floor := range numFloors {
		for btn := range numBtns - 1 {
			elev.LocalOrders[floor][btn] = false
		}
	}
	return elev
}

func setCabLights(elev elevator.Elevator) {

	for floor := 0; floor < numFloors; floor++ {
		btn := int(elevio.BT_Cab)
		if elev.LocalOrders[floor][btn] {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
		} else {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
}


