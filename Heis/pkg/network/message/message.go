package message

import (
	"Heis/pkg/elevator"
)

type ElevatorStateMsg struct {
	// Message        string
	Iter            int
	NetworkElevator elevator.NetworkElevator
	Id              string
}
