package types

import "Heis/pkg/elevio"

type UdpMsg struct {
	Message        string
	Iter           int
	ButtonPressMsg *ButtonPressMsg
	ClearFloorMsg  *ClearFloorMsg
}

type ButtonPressMsg struct {
	Floor  int
	Button elevio.ButtonType
	Id     string
}

type ClearFloorMsg struct {
	Floor int
	// Should also include the direction to clear.
	Dirn elevio.MotorDirection
	Id   string
}
