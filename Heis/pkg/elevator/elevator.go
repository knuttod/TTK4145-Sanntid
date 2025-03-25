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
	Ordr_None        OrderState = 0
	Ordr_Unconfirmed OrderState = 1
	Ordr_Confirmed   OrderState = 2
	Ordr_Complete    OrderState = 3

	Ordr_Unknown OrderState = -1
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
	Behaviour   ElevatorBehaviour
	Obstructed  bool
	MotorStop   bool
	LocalOrders [][]bool

	Id     string
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
func Elevator_init(N_floors, N_buttons int, id string) Elevator {
	// initialize the (*e) struct
	var elev Elevator
	elev.Floor = -1
	elev.Dirn = elevio.MD_Stop
	elev.Behaviour = EB_Idle
	elev.Obstructed = false
	elev.MotorStop = false

	elev.LocalOrders = make([][]bool, N_floors)
	for floor := range N_floors {
		elev.LocalOrders[floor] = make([]bool, N_buttons)
		for btn := range N_buttons {
			elev.LocalOrders[floor][btn] = false
		}
	}

	elev.Id = id

	elev.Config.ClearRequestVariant = CV_InDirn
	elev.Config.DoorOpenDuration_s = 3.0

	return elev
}
