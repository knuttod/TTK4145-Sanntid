package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/elevator"
	//"fmt"
)

// type DirnBehaviourPair struct {
// 	Behaviour ElevatorBehaviour
// 	Dirn      elevio.MotorDirection
// }

func Above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < N_floors; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.LocalOrders[f][btn] == 2 {
				return true
			}
		}
	}

	return false
}

func Below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.LocalOrders[f][btn] == 2 {
				return true
			}
		}
	}
	return false
}

func Here(e elevator.Elevator) bool {
	for btn := 0; btn < N_buttons; btn++ {
		if e.LocalOrders[e.Floor][btn] == 2{
			return true
		}
	}
	return false
}

func chooseDirection(e elevator.Elevator) elevator.DirnBehaviourPair {

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
		if (e.LocalOrders[e.Floor][elevio.BT_HallDown] == 2) || (e.LocalOrders[e.Floor][elevio.BT_Cab] == 2) || (!Below(e)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Up:
		if (e.LocalOrders[e.Floor][elevio.BT_HallUp] == 2) || (e.LocalOrders[e.Floor][elevio.BT_Cab] == 2) || (!Above(e)) {
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


// Få inn funksjonaliltet for å sende eller opdatere globale orders når denne blir kjørt
func ClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		for btn := 0; btn < N_buttons; btn++ {
			e.LocalOrders[e.Floor][btn] = 0
		}
	case elevator.CV_InDirn:
		e.LocalOrders[e.Floor][elevio.BT_Cab] = 0
		switch e.Dirn {
		case elevio.MD_Up:
			if (!Above(e)) && (e.LocalOrders[e.Floor][elevio.BT_HallUp] != 2) {
				e.LocalOrders[e.Floor][elevio.BT_HallDown] = 0
			}
			e.LocalOrders[e.Floor][elevio.BT_HallUp] = 0

		case elevio.MD_Down:
			if (!Below(e)) && (e.LocalOrders[e.Floor][elevio.BT_HallDown] != 2) {
				e.LocalOrders[e.Floor][elevio.BT_HallUp] = 0
			}
			e.LocalOrders[e.Floor][elevio.BT_HallDown] = 0
		// case elevio.MD_Stop:
		// 	e.LocalOrders[e.Floor][elevio.BT_HallUp] = 0
		// 	e.LocalOrders[e.Floor][elevio.BT_HallDown] = 0
		// 	e.LocalOrders[e.Floor][elevio.BT_Cab] = 0
		default:
			e.LocalOrders[e.Floor][elevio.BT_HallUp] = 0
			e.LocalOrders[e.Floor][elevio.BT_HallDown] = 0
			//e.LocalOrders[e.Floor][elevio.BT_Cab] = 0
		}
	default:

	}
	return e
}

// // Funker ikke som den skal den fjerner alt på samme etasje, tar ikke hensyn til rettning for øyeblikket.
// func ClearFloor(e *types.Elevator, floor int, Dirn elevio.MotorDirection) {
// 	switch Dirn {
// 	case elevio.MD_Up:
// 		if (!Above(e)) && (!e.Requests[e.Floor][elevio.BT_HallUp]) { //trenger en workaround for Above()
// 			e.Requests[floor][elevio.BT_HallDown] = false
// 		}
// 		e.Requests[floor][elevio.BT_HallUp] = false

// 	case elevio.MD_Down:
// 		if (!Below(e)) && (!e.Requests[e.Floor][elevio.BT_HallDown]) { // trenger en workaround for Below() også
// 			e.Requests[floor][elevio.BT_HallUp] = false
// 		}
// 		e.Requests[floor][elevio.BT_HallDown] = false
// 	// case elevio.MD_Stop:
// 	// 	e.Requests[e.Floor][elevio.BT_HallUp] = false
// 	// 	e.Requests[e.Floor][elevio.BT_HallDown] = false
// 	// 	e.Requests[e.Floor][elevio.BT_Cab] = false
// 	default:
// 		e.Requests[floor][elevio.BT_HallUp] = false
// 		e.Requests[floor][elevio.BT_HallDown] = false
// 		//e.Requests[e.Floor][elevio.BT_Cab] = false
// 	}

// 	setAllLights(e)
// }
