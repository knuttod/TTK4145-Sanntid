package fsm

import (
	"Heis/pkg/elevio"
)

func fsm_init() {
	// initialize the Elevator struct
	elevator.Floor = -1
	elevator.Dirn = elevio.MD_Stop
	elevator.Behaviour = EB_Idle
	elevator.Config.ClearRequestVariant = CV_InDirn
	elevator.Config.DoorOpenDuration_s = 3.0
}

func initBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = EB_Moving
}

func requestButtonPress(btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64) {
	//print functions??

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if requests.shouldClearImmediately(elevator, btn_floor, btn_type) {
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
		} else {
			elevator.requests[btn_floor][btn_type] = 1
		}

	case EB_Moving:
		elevator.requests[btn_floor][btn_type] = 1

	case EB_Idle:
		elevator.requests[btn_floor][btn_type] = 1
		var pair requests.DirnBehaviourPair = requests.chooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
			elevator = requests.clearAtCurrentFloor(elevator)

		case EB_Moving:
			elevio.SetMotorDirection(elevator.Dirn)

			//case EB_Idle:
		}

	}

	setAllLights(elevator)
}

func floorArrival(newFloor int, drv_doorTimer chan float64) {
	elevator.Floor = newFloor
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case EB_Moving:
		if requests.shouldStop(elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = requests.clearAtCurrentFloor(elevator)
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
			setAllLights(elevator)
			elevator.Behaviour = EB_DoorOpen
		}
	}
}

func DoorTimeout(drv_doorTimer chan float64) {

	switch elevator.Behaviour {
	case EB_DoorOpen:
		var pair requests.DirnBehaviourPair = requests.chooseDirection(elevator)
		elevator.Dirn = pair.Dirn
		elevator.Behaviour = pair.Behaviour

		switch elevator.Behaviour {
		case EB_DoorOpen:
			drv_doorTimer <- elevator.Config.DoorOpenDuration_s
			elevator = requests.clearAtCurrentFloor(elevator)
			setAllLights(elevator)

		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevator.Dirn)
		}
	}
}

func setAllLights(e Elevator) {
	//set ligths
	for 
}
