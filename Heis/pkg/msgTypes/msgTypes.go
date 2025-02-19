package msgTypes

import (
	"Heis/pkg/elevio"
	"Heis/pkg/elevator"
)


type UdpMsg struct {
	Message        string
	Iter           int
	ElevatorStateMsg *ElevatorStateMsg // Pointer to ButtonPressMsg, nil if not a button press
}

type ButtonPressMsg struct {
	Floor  int
	Button elevio.ButtonType
	Id     string // Identifier of the elevator that pressed the button
}

type ElevatorStateMsg struct {
	Elevator *elevator.Elevator
	Id       string
}

