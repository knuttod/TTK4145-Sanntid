package fsm

import (
	"Heis/pkg/elevio"
)

func fsm_init() {
	// initialize the Elevator struct
	elevator.Floor = -1
	elevator.dirn = elevio.MD_Stop
	elevator.behaviour = EB_Idle
	elevator.config.clearRequestVariant = CV_InDirn
	elevator.config.doorOpenDuration_s = 3.0
}

func initBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.dirn = elevio.MD_Down
	elevator.behaviour = EB_Moving
}

func requestButtonPress(btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64) {
	//print functions??

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests.shouldClearImmediately(elevator, btn_floor, btn_type) {
			drv_doorTimer <- elevator.config.doorOpenDuration_s
		} else {
			elevator.requests[btn_floor][btn_type] = 1
		}

	case EB_Moving:
		elevator.requests[btn_floor][btn_type] = 1

	case EB_Idle:
		elevator.requests[btn_floor][btn_type] = 1
		var pair requests.DirnBehaviourPair = requests.chooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour

		switch pair.behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			drv_doorTimer <- elevator.config.doorOpenDuration_s
			elevator = requests.clearAtCurrentFloor(elevator)

		case EB_Moving:
			elevio.SetMotorDirection(elevator.dirn)

			//case EB_Idle:
		}

	}

	setAllLights(elevator)
}

func floorArrival(newFloor int, drv_doorTimer chan float64) {
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.behaviour {
	case EB_Moving:
		if requests.shouldStop(elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = requests.clearAtCurrentFloor(elevator)
			drv_doorTimer <- elevator.config.doorOpenDuration_s
			setAllLights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
	}
}

func DoorTimeout(drv_doorTimer chan float64) {

	switch elevator.behaviour {
	case EB_DoorOpen:
		var pair requests.DirnBehaviourPair = requests.chooseDirection(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour

		switch elevator.behaviour {
		case EB_DoorOpen:
			drv_doorTimer <- elevator.config.doorOpenDuration_s
			elevator = requests.clearAtCurrentFloor(elevator)
			setAllLights(elevator)

		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.dirn)
		}
	}
}

func setAllLights(e Elevator) {
	//set ligths
}
