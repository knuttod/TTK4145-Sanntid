package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	//"fmt"
)

func Above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < N_floors; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.LocalOrders[f][btn] == true {
				return true
			}
		}
	}

	return false
}

func Below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.LocalOrders[f][btn] == true {
				return true
			}
		}
	}
	return false
}

func Here(e elevator.Elevator) bool {
	for btn := 0; btn < N_buttons; btn++ {
		if e.LocalOrders[e.Floor][btn] == true {
			return true
		}
	}
	return false
}

func ChooseDirection(e elevator.Elevator) elevator.DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if Above(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else if Here(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_DoorOpen}
		} else if Below(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}
	case elevio.MD_Down:
		if Below(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else if Here(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_DoorOpen}
		} else if Above(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if Here(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_DoorOpen}
		} else if Above(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: elevator.EB_Moving}
		} else if Below(e) {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: elevator.EB_Moving}
		} else {
			return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
		}
	default:
		return elevator.DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: elevator.EB_Idle}
	}

}

func ShouldStop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		if (e.LocalOrders[e.Floor][elevio.BT_HallDown] == true) || (e.LocalOrders[e.Floor][elevio.BT_Cab] == true) || (!Below(e)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Up:
		if (e.LocalOrders[e.Floor][elevio.BT_HallUp] == true) || (e.LocalOrders[e.Floor][elevio.BT_Cab] == true) || (!Above(e)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Stop:
	default:
		return true

	}
	return false
}

func ShouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		if e.Floor == btn_floor {
			return true
		} else {
			return false
		}
	case elevator.CV_InDirn:
		if e.Floor == btn_floor && ((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			(e.Dirn == elevio.MD_Stop) || (btn_type == elevio.BT_Cab)) {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}


// Clears at current floor and sends that the order is complete to the order module.
func ClearAtCurrentFloor(e elevator.Elevator, completedOrder chan elevio.ButtonEvent) elevator.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		for btn := 0; btn < N_buttons; btn++ {
			e.LocalOrders[e.Floor][btn] = false
			completedOrder <- elevio.ButtonEvent {
				Floor: e.Floor,
				Button: elevio.ButtonType(btn),
			}
		}
	case elevator.CV_InDirn:
		e.LocalOrders[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			if (!Above(e)) && (e.LocalOrders[e.Floor][elevio.BT_HallUp] == false) {
				e.LocalOrders[e.Floor][elevio.BT_HallDown] = false
				completedOrder <- elevio.ButtonEvent {
					Floor: e.Floor,
					Button: elevio.BT_HallDown,
				}
			}
			e.LocalOrders[e.Floor][elevio.BT_HallUp] = false
			completedOrder <- elevio.ButtonEvent {
				Floor: e.Floor,
				Button: elevio.BT_HallUp,
			}

		case elevio.MD_Down:
			if (!Below(e)) && (e.LocalOrders[e.Floor][elevio.BT_HallDown] == false) {
				e.LocalOrders[e.Floor][elevio.BT_HallUp] = false
				completedOrder <- elevio.ButtonEvent {
					Floor: e.Floor,
					Button: elevio.BT_HallUp,
				}
			}
			e.LocalOrders[e.Floor][elevio.BT_HallDown] = false
			completedOrder <- elevio.ButtonEvent {
				Floor: e.Floor,
				Button: elevio.BT_HallDown,
			}
		// case elevio.MD_Stop:
		// 	e.LocalOrders[e.Floor][elevio.BT_HallUp] = false
		// 	e.LocalOrders[e.Floor][elevio.BT_HallDown] = false
		// 	e.LocalOrders[e.Floor][elevio.BT_Cab] = false
		default:
			e.LocalOrders[e.Floor][elevio.BT_HallUp] = false
			e.LocalOrders[e.Floor][elevio.BT_HallDown] = false
			//e.LocalOrders[e.Floor][elevio.BT_Cab] = false
			completedOrder <- elevio.ButtonEvent {
				Floor: e.Floor,
				Button: elevio.BT_HallDown,
			}
			completedOrder <- elevio.ButtonEvent {
				Floor: e.Floor,
				Button: elevio.BT_HallUp,
			}
		}
	default:

	}
	return e
}
