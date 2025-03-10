package msgTypes

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
)


type UdpMsg struct {
	Message        string
	Iter           int
	ElevatorStateMsg *ElevatorStateMsg // Pointer to ButtonPressMsg, nil if not a button press
}


type ElevatorStateMsg struct {
	Message        string
	Iter           int
	NetworkElevator elevator.NetworkElevator
	Id       string
}


type FsmMsg struct {
	Elevator 	elevator.Elevator
	Event 		elevio.ButtonEvent
}

