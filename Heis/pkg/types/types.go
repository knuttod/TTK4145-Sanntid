package types

import "Heis/pkg/elevio"

type UdpMsg struct {
	Message        string
	Iter           int
	ButtonPressMsg *ButtonPressMsg // Pointer to ButtonPressMsg, nil if not a button press
}

type ButtonPressMsg struct {
	Floor  int
	Button elevio.ButtonType
	Id     string // Identifier of the elevator that pressed the button
}
