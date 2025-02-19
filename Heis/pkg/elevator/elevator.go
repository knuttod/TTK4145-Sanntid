package elevator

import "Heis/pkg/elevio"



type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_ALL ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	Floor      int
	Dirn       elevio.MotorDirection
	Requests   [][]bool
	Behaviour  ElevatorBehaviour
	Obstructed bool
	Id string

	Config struct { //type?
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}

type DirnBehaviourPair struct {
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
}


func Elevator_init(e *Elevator, N_floors, N_buttons int, id string) {
	// initialize the (*e) struct
	(*e).Floor = -1
	(*e).Dirn = elevio.MD_Stop
	(*e).Behaviour = EB_Idle
	(*e).Config.ClearRequestVariant = CV_InDirn
	(*e).Config.DoorOpenDuration_s = 3.0
	(*e).Id = id
	(*e).Requests = make([][]bool, N_floors)
	for i := range (*e).Requests {
		(*e).Requests[i] = make([]bool, N_buttons)
	}
}

// func transmitState(e *Elevator, Tx chan msgTypes.UdpMsg, id string) Elevator {
// 	elevatorStateMsg := msgTypes.ElevatorStateMsg{
// 			Elevator: e,
// 			Id:       id,
// 		}
// 	for {
// 		Tx <- msgTypes.UdpMsg{ButtonPressMsg: &buttonPressMsg}
// 		time.Sleep(10 * time.Millisecond)
// 	}
// }

