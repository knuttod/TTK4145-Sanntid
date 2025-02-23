package elevator

import (
	"Heis/pkg/elevio"
	"strconv"
)





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
	LocalOrders   [][]int
	Behaviour  ElevatorBehaviour
	Obstructed bool
	Id 		   string
	Index 		   int
	GlobalOrders 	   [][]int	//Acts as a cyclic counter with 0 completed/no order, 1 unconfirmed order, 2 confirmed order. Two first rows are hall orders, then there are cab calls for elev1, elev2 and so on

	Config struct { //type?
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}

type Order struct {
	State int
	Action elevio.ButtonEvent 
}

type DirnBehaviourPair struct {
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
}


func Elevator_init(e *Elevator, N_floors, N_buttons, N_elevators int, id string) {
	// initialize the (*e) struct
	(*e).Floor = -1
	(*e).Dirn = elevio.MD_Stop
	(*e).Behaviour = EB_Idle
	(*e).Config.ClearRequestVariant = CV_InDirn
	(*e).Config.DoorOpenDuration_s = 3.0
	(*e).Id = id
	//Assuming an Id/index from 1 - 9
	i, err := strconv.Atoi(id[len(id) - 1:])
	if err != nil {
		(*e).Index = 0
	} else {
		(*e).Index = i
	}
	
	(*e).LocalOrders = make([][]int, N_floors)
	for i := range (*e).LocalOrders {
		(*e).LocalOrders[i] = make([]int, N_buttons)
	}
	(*e).GlobalOrders = make([][]int, N_floors)
	for i := range (*e).GlobalOrders {
		(*e).GlobalOrders[i] = make([]int, 2 + N_elevators)
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

