package fsm

import (
	"Heis/pkg/elevio"
)

func fsm_init(e *Elevator) {
	// initialize the Elevator struct
	elevator := *e
	elevator.Floor = -1
	elevator.Dirn = elevio.MD_Stop
	elevator.Behaviour = EB_Idle
	elevator.Config.ClearRequestVariant = CV_InDirn
	elevator.Config.DoorOpenDuration_s = 3.0
}

func initBetweenFloors(e *Elevator) {
	elevator := *e
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = EB_Moving
}

func requestButtonPress(e *Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64) {
	//print functions??
	elevator := *e

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if ShouldClearImmediately(elevator, btn_floor, btn_type) {
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
		} else {
			elevator.Requests[btn_floor][btn_type] = 1
		}

	case EB_Moving:
		elevator.Requests[btn_floor][btn_type] = 1

	case EB_Idle:
		elevator.Requests[btn_floor][btn_type] = 1
		var pair DirnBehaviourPair = chooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
			elevator = ClearAtCurrentFloor(elevator)

		case EB_Moving:
			elevio.SetMotorDirection(elevator.Dirn)

			//case EB_Idle:
		}

	}
	//setAllLights(elevator)
}

func floorArrival(e *Elevator, newFloor int, drv_doorTimer chan float64) {
	elevator := *e

	elevator.Floor = newFloor
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case EB_Moving:
		if ShouldStop(elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = ClearAtCurrentFloor(elevator)
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
			//setAllLights(elevator)
			elevator.Behaviour = EB_DoorOpen
		}
	}
}

func DoorTimeout(e *Elevator, drv_doorTimer chan float64) {
	elevator := *e

	switch elevator.Behaviour {
	case EB_DoorOpen:
		var pair DirnBehaviourPair = chooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case EB_DoorOpen:
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
			elevator = ClearAtCurrentFloor(elevator)
			//setAllLights(elevator)

		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)

		}
	}
}

// func setAllLights(e Elevator) {
// 	//set ligths
// 	for floor := 0; floor < N_floors; floor++ {
// 		for btn := 0; btn < N_buttons; btn++ {

// 		}
// 	}
// }
