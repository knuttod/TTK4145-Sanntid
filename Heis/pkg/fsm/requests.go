package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/types"
)

type DirnBehaviourPair struct {
	Behaviour types.ElevatorBehaviour
	Dirn      elevio.MotorDirection
}

func Above(e types.Elevator) bool {
	for f := e.Floor + 1; f < N_floors; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}

	return false
}

func Below(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func Here(e types.Elevator) bool {
	for btn := 0; btn < N_buttons; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func chooseDirection(e types.Elevator) DirnBehaviourPair {

	switch e.Dirn {
	case elevio.MD_Up:
		if Above(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: types.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: types.EB_DoorOpen}
		} else if Below(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: types.EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
		}
	case elevio.MD_Down:
		if Below(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: types.EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: types.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: types.EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
		}

	case elevio.MD_Stop:
		if Here(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: types.EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: types.EB_Moving}
		} else if Below(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: types.EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
	}

}

func ShouldStop(e types.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		if (e.Requests[e.Floor][elevio.BT_HallDown]) || (e.Requests[e.Floor][elevio.BT_Cab]) || (!Below(e)) {
			return true
		} else {
			return false
		}

	case elevio.MD_Up:
		if (e.Requests[e.Floor][elevio.BT_HallUp]) || (e.Requests[e.Floor][elevio.BT_Cab]) || (!Above(e)) {
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

func ShouldClearImmediately(e types.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case types.CV_ALL:
		if e.Floor == btn_floor {
			return true
		} else {
			return false
		}
	case types.CV_InDirn:
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

func ClearAtCurrentFloor(e types.Elevator) types.Elevator {
	switch e.Config.ClearRequestVariant {
	case types.CV_ALL:
		for btn := 0; btn < N_buttons; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case types.CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			if (!Above(e)) && (!e.Requests[e.Floor][elevio.BT_HallUp]) {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}

		case elevio.MD_Down:
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			if (!Below(e)) && (!e.Requests[e.Floor][elevio.BT_HallDown]) {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
		// case elevio.MD_Stop:
		// 	e.Requests[e.Floor][elevio.BT_HallUp] = false
		// 	e.Requests[e.Floor][elevio.BT_HallDown] = false
		// 	e.Requests[e.Floor][elevio.BT_Cab] = false
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			//e.Requests[e.Floor][elevio.BT_Cab] = false
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
