package fsm

import "Heis/pkg/elevio"

func fsm_init(elevator *Elevator) (Elevator, bool) {
	// initialize the Elevator struct
}

func initBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)

}
