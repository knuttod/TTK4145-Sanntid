package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	//"fmt"
)

// When initialising between floors, the motor direction is set downwards.
// When it reaches a floor it will then behave normaly.
func initBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = elevator.EB_Moving
}

// Handles button presses on a local level, by processing requests based on the
// elevator's current behavior. If the elevator is idle, it determines the next action
// (moving or opening doors). If the elevator is moving or has doors open, it updates
// the request state accordingly. The function also manages the door timer, sends updated
// elevator states over UDP, and updates the button lights.
func requestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64, Tx chan msgTypes.UdpMsg, id string) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
		} else {
			(*e).Requests[btn_floor][btn_type] = true
		}

	case elevator.EB_Moving:
		(*e).Requests[btn_floor][btn_type] = true

	case elevator.EB_Idle:
		(*e).Requests[btn_floor][btn_type] = true
		var pair elevator.DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			//drv_doorTimer <- 0.0
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			(*e) = ClearAtCurrentFloor((*e))

		case elevator.EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
			//clear something at this floor??

		case elevator.EB_Idle:
			//need something here?
		}

	}
	setAllLights(e)
}

// When arriving at a floor this sets the floor indicator to the floor, and checks if it is supposed
// to stop. if it is supposed to stop it stops, clears the floor then opens the door.
func floorArrival(e *elevator.Elevator, newFloor int, drv_doorTimer chan float64) {

	(*e).Floor = newFloor
	elevio.SetFloorIndicator((*e).Floor)

	switch (*e).Behaviour {
	case elevator.EB_Moving:
		if ShouldStop((*e)) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			(*e) = ClearAtCurrentFloor((*e))
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
			setAllLights(e)
			(*e).Behaviour = elevator.EB_DoorOpen
		}
	}
}

// DoorTimeout is used to start the timer when the door openes, and
// handles if there is an obstruction when the door closes. It is
// runned twice, once at the begining of the timer initialisation and
// once when the door is supposed to close to check if the obstruction
// is active.
func DoorTimeout(e *elevator.Elevator, drv_doorTimer chan float64) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		var pair elevator.DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch (*e).Behaviour {
		case elevator.EB_DoorOpen:
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s //????
			//drv_doorTimer <- 0.0
			(*e) = ClearAtCurrentFloor((*e))
			setAllLights(e)

		//lagt inn selv
		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)
		//

		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)

		}
	}
}

// Updates the elevator's button lights to reflect the current request states.
// Turns on lights for active requests and turns them off otherwise.
func setAllLights(e *elevator.Elevator) {
	//set ligths
	for floor := 0; floor < N_floors; floor++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[floor][btn] {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}
