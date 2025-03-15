package elevator

import (
	"Heis/pkg/elevio"
)

type ClearRequestVariant int

const (
	CV_ALL ClearRequestVariant = iota
	CV_InDirn
)

type OrderState int

const (
	Ordr_None 			OrderState = 0
	Ordr_Unconfirmed    OrderState = 1
	Ordr_Confirmed 		OrderState = 2
	Ordr_Complete		OrderState = 3
	
	Ordr_Unknown   		OrderState = -1
)

type ElevatorBehaviour int

const (
	EB_Idle        ElevatorBehaviour = 0
	EB_DoorOpen    ElevatorBehaviour = 1
	EB_Moving      ElevatorBehaviour = 2
	EB_Unavailable ElevatorBehaviour = 3
)

type Elevator struct {
	Floor       int
	Dirn        elevio.MotorDirection
	LocalOrders [][]bool
	// Map er egentlig litt lite effektivt da man ikke kan opdatere deler av verdien på en key, men heller må bare gi en helt ny verdi/struct.
	// AssignedOrders   map[string][][]OrderState
	Behaviour  ElevatorBehaviour
	Obstructed bool
	Id         string

	Config struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}

type NetworkElevator struct {
	Elevator       Elevator
	AssignedOrders map[string][][]OrderState
}

type DirnBehaviourPair struct {
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
}

// Initializes an elevator struct. All orders are by default set to 0/false
func Elevator_init(e *Elevator, N_floors, N_buttons, N_elevators int, id string) {
	// initialize the (*e) struct
	(*e).Floor = -1
	(*e).Dirn = elevio.MD_Stop
	(*e).Behaviour = EB_Idle
	(*e).Config.ClearRequestVariant = CV_InDirn
	(*e).Config.DoorOpenDuration_s = 3.0
	(*e).Id = id

	(*e).LocalOrders = make([][]bool, N_floors)
	for i := range (*e).LocalOrders {
		(*e).LocalOrders[i] = make([]bool, N_buttons)
	}
}
