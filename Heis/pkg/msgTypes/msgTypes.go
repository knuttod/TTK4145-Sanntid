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
	Message        string
	Iter           int
	Floor  		   int
	Button elevio.ButtonType
	Id     string // Identifier of the elevator that pressed the button
}

type ElevatorStateMsg struct {
	Message        string
	Iter           int
	Elevator elevator.Elevator
	Id       string
}

