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

type RequestState int

const (
	None      RequestState = 0
	Order     RequestState = 1
	Comfirmed RequestState = 2
	Complete  RequestState = 3
)

type Behaviour int

const (
	Idle        Behaviour = 0
	DoorOpen    Behaviour = 1
	Moving      Behaviour = 2
	Unavailable Behaviour = 3
)

type Elevator struct {
	Floor      int
	Dirn       elevio.MotorDirection
	Requests   [][]RequestState		//acts as cyclic counter
	AssignedOrders   map[string]([N_floors][N_buttons]RequestState)
	Behaviour  ElevatorBehaviour
	Obstructed bool
	Id 		   string
	Index 	   int
	//ExternalElevators []Elevator	//Dette funker ikke, vet ikke hvorfor

	Config struct { //type?
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}

// type DistributorElevator struct {
// 	ID       string
// 	Floor    int
// 	Dir      Direction
// 	Requests [][]RequestState
// 	Behave   Behaviour
// }

type Request struct {
	Floor  int
	Button elevio.ButtonType
}

type CostRequest struct {
	Id         string
	Cost       int
	AssignedId string
	Req        Request
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
	
	AssignedOrders := make([][]int, N_floors)
	for i := range AssignedOrders {
		AssignedOrders[i] = make([]int, N_buttons)
	}
	(e*).AssignedOrders = map[string]([N_floors][N_buttons]RequestState){id: AssginedOrders}

	(*e).Requests = make([][]int, N_floors)
	for i := range (*e).Requests {
		(*e).Requests[i] = make([]int, 2 + N_elevators)
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

