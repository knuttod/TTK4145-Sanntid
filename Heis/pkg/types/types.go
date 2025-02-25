package types

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

	Config struct { //type?
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}
type UdpMsg struct {
	Message          string
	Iter             int
	ButtonPressMsg   *ButtonPressMsg
	ClearFloorMsg    *ClearFloorMsg
	ElevatorStateMsg *ElevatorStateMsg
}

type ElevatorStateMsg struct {
	Elevator Elevator
	Id       string
}

type ButtonPressMsg struct {
	Floor  int
	Button elevio.ButtonType
	Id     string
}

type ClearFloorMsg struct {
	Floor int
	Dirn  elevio.MotorDirection
	Id    string
}
