package elevator

import (
	"Heis/pkg/elevio"
	"strconv"
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
	Confirmed RequestState = 2
	Complete  RequestState = 3
)

type ElevatorBehaviour int

const (
	EB_Idle        ElevatorBehaviour = 0
	EB_DoorOpen    ElevatorBehaviour = 1
	EB_Moving      ElevatorBehaviour = 2
	EB_Unavailable ElevatorBehaviour = 3
)

type Elevator struct {
	Floor      int
	Dirn       elevio.MotorDirection
	Orders   [][]RequestState		//acts as cyclic counter
	// Map er egentlig litt lite effektivt da man ikke kan opdatere deler av verdien på en key, men heller må bare gi en helt ny verdi/struct. 
	AssignedOrders   map[string][][]RequestState
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
	
	AssignedOrders := make([][]RequestState, N_floors)
	for i := range AssignedOrders {
		AssignedOrders[i] = make([]RequestState, N_buttons)
	}
	(*e).AssignedOrders = map[string][][]RequestState{id: AssignedOrders}

	(*e).Orders = make([][]RequestState, N_floors)
	for i := range (*e).Orders {
		(*e).Orders[i] = make([]RequestState, N_buttons)
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

