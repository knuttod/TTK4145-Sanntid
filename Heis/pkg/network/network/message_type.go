package network

import (
	"Heis/pkg/elevator"
)

type ElevatorStateMsg struct {
	NetworkElevator elevator.NetworkElevator
	Id              string
}
