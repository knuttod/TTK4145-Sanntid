package request

import (
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
)

type DirnBehaviourPair struct {
	Behaviour fsm.ElevatorBehaviour
	Dirn      elevio.MotorDirection
}

func Above(e fsm.Elevator) int {
	for f := e.Floor; f < fsm.N_floors; f++ {
		for btn := 0; btn < fsm.N_buttons; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}

	return 0
}

func Below(e fsm.Elevator) int {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < fsm.N_buttons; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func Here(e fsm.Elevator) int {
	for btn := 0; btn < fsm.N_buttons; btn++ {
		if e.Requests[e.Floor][btn] != 0 {
			return 1
		}
	}
	return 0
}

func chooseDirection(e fsm.Elevator) DirnBehaviourPair {

	switch e.Dirn {
	case elevio.MD_Up:
		if Above(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: fsm.EB_Moving}
		} else if Here(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: fsm.EB_DoorOpen}
		} else if Below(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: fsm.EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: fsm.EB_Idle}
		}
	case elevio.MD_Down:
		if Below(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: fsm.EB_Moving}
		} else if Here(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: fsm.EB_DoorOpen}
		} else if Above(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: fsm.EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: fsm.EB_Idle}
		}

	case elevio.MD_Stop:
		if Here(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: fsm.EB_DoorOpen}
		} else if Above(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Up, Behaviour: fsm.EB_Moving}
		} else if Below(e) != 0 {
			return DirnBehaviourPair{Dirn: elevio.MD_Down, Behaviour: fsm.EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: fsm.EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: elevio.MD_Stop, Behaviour: fsm.EB_Idle}
	}

}

func ShouldStop(e fsm.Elevator) int {
	switch e.Dirn {
	case elevio.MD_Down:
		if (e.Requests[e.Floor][elevio.BT_HallDown] != 0) || (e.Requests[e.Floor][elevio.BT_Cab] != 0) || (Below(e) == 0) {
			return 1
		} else {
			return 0
		}

	case elevio.MD_Up:
		if (e.Requests[e.Floor][elevio.BT_HallUp] != 0) || (e.Requests[e.Floor][elevio.BT_Cab] != 0) || (Below(e) == 0) {
			return 1
		} else {
			return 0
		}

	case elevio.MD_Stop:
	default:
		return 1

	}
	return 0
}

func ShouldClearImmediately(e fsm.Elevator, btn_floor int, btn_type elevio.ButtonType) int {
	switch e.Config.ClearRequestVariant {
	case fsm.CV_ALL:
		if e.Floor == btn_floor {
			return 1
		} else {
			return 0
		}
	case fsm.CV_InDirn:
		if e.Floor == btn_floor && ((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			(e.Dirn == elevio.MD_Stop) || (btn_type == elevio.BT_Cab)) {
			return 1
		} else {
			return 0
		}
	default:
		return 0
	}
}

func ClearAtCurrentFloor(e fsm.Elevator) fsm.Elevator {
	switch e.Config.ClearRequestVariant {
	case fsm.CV_ALL:
		for btn := 0; btn < fsm.N_buttons; btn++ {
			e.Requests[e.Floor][btn] = 0
		}
	case fsm.CV_InDirn:
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
