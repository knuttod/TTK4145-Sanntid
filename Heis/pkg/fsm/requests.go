package fsm

import (
	"Heis/pkg/elevio"
)

type DirnBehaviourPair struct {
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
}

func Above(e Elevator) int {
	for f := e.Floor; f < N_floors; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}

	return 0
}

func Below(e Elevator) int {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func Here(e Elevator) int {
	for btn := 0; btn < N_buttons; btn++ {
		if e.Requests[e.Floor][btn] != 0 {
			return 1
		}
	}
	return 0
}

func chooseDirection(e Elevator) DirnBehaviourPair {

	switch e.Dirn {
	case elevio.MD_Up:
		if Above(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_Moving}
		} else if Here(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_DoorOpen}
		} else if Below(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_Idle}
		}
	case elevio.MD_Down:
		if Below(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: EB_Moving}
		} else if Here(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_DoorOpen}
		} else if Above(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_Idle}
		}

	case elevio.MD_Stop:
		if Here(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: EB_DoorOpen}
		} else if Above(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: EB_Moving}
		} else if Below(e) != 0 {
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
		if (e.Requests[e.Floor][elevio.BT_HallDown] != 0) || (e.Requests[e.Floor][elevio.BT_Cab] != 0) || (Below(e) == 0) {
			return true
		} else {
			return false
		}

	case elevio.MD_Up:
		if (e.Requests[e.Floor][elevio.BT_HallUp] != 0) || (e.Requests[e.Floor][elevio.BT_Cab] != 0) || (Below(e) == 0) {
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
			e.Requests[e.Floor][btn] = 0
		}
	case CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = 0
		switch e.Dirn {
		case elevio.MD_Up:
			if Above(e) == 0 && (e.Requests[e.Floor][elevio.BT_HallUp] == 0) {
				e.Requests[e.Floor][elevio.BT_HallDown] = 0
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = 0

		case elevio.MD_Down:
			if (Below(e) == 0) && (e.Requests[e.Floor][elevio.BT_HallDown] == 0) {
				e.Requests[e.Floor][elevio.BT_HallUp] = 0
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = 0
		case elevio.MD_Stop:
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = 0
			e.Requests[e.Floor][elevio.BT_HallDown] = 0
		}
	default:

	}
	return e
}
