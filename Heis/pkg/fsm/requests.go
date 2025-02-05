package fsm

import (
	"Heis/pkg/elevio"
	//"fmt"
)

type DirnBehaviourPair struct {
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
}

func Above(e Elevator) bool {
	for f := e.Floor + 1; f < N_floors; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}

	return false
}

func Below(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func Here(e Elevator) bool {
	for btn := 0; btn < N_buttons; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func chooseDirection(e Elevator) DirnBehaviourPair {

	switch e.Dirn {
	case elevio.MD_Up:
		if Above(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_DoorOpen}
		} else if Below(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_Idle}
		}
	case elevio.MD_Down:
		if Below(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_Moving}
		} else if Here(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_Idle}
		}

	case elevio.MD_Stop:
		if Here(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_DoorOpen}
		} else if Above(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_Moving}
		} else if Below(e) {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_Idle}
	}

}

func ShouldStop(e Elevator) bool {
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

func ShouldClearImmediately(e Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case CV_ALL:
		if e.Floor == btn_floor {
			return true
		} else {
			return false
		}
	case CV_InDirn:
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

func ClearAtCurrentFloor(e Elevator) Elevator {
	switch e.Config.ClearRequestVariant {
	case CV_ALL:
		for btn := 0; btn < N_buttons; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			if (!Above(e)) && (!e.Requests[e.Floor][elevio.BT_HallUp]) {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = false

		case elevio.MD_Down:
			if (!Below(e)) && (!e.Requests[e.Floor][elevio.BT_HallDown]) {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		//case elevio.MD_Stop:
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
	default:

	}
	return e
}
