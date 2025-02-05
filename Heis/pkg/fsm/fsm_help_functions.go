package fsm

import (
	"Heis/pkg/elevio"
	"fmt"
)

func fsm_init(e *Elevator) {
	// initialize the (*e) struct
	(*e).Floor = -1
	(*e).Dirn = elevio.MD_Stop
	(*e).Behaviour = EB_Idle
	(*e).Config.ClearRequestVariant = CV_InDirn
	(*e).Config.DoorOpenDuration_s = 3.0
	(*e).Requests = make([][]bool, N_floors)
	for i := range (*e).Requests {
		(*e).Requests[i] = make([]bool, N_buttons)
	}
}

func initBetweenFloors(e *Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = EB_Moving
}

func requestButtonPress(e *Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64) {
	//print functions??

	switch (*e).Behaviour {
	case EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
		} else {
			(*e).Requests[btn_floor][btn_type] = true
		}

	case EB_Moving:
		(*e).Requests[btn_floor][btn_type] = true

	case EB_Idle:
		(*e).Requests[btn_floor][btn_type] = true
		var pair DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			(*e) = ClearAtCurrentFloor((*e))

		case EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)

			//case EB_Idle:
		}

	}
	setAllLights(e)
}

func floorArrival(e *Elevator, newFloor int, drv_doorTimer chan float64) {

	fmt.Println("Floor")

	(*e).Floor = newFloor
	elevio.SetFloorIndicator((*e).Floor)

	switch (*e).Behaviour {
	case EB_Moving:
		fmt.Println("liten stop")
		if ShouldStop((*e)) {
			fmt.Println("STOP!!!!")
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			(*e) = ClearAtCurrentFloor((*e))
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
			setAllLights(e)
			(*e).Behaviour = EB_DoorOpen
		}
	}
}

func DoorTimeout(e *Elevator, drv_doorTimer chan float64) {

	switch (*e).Behaviour {
	case EB_DoorOpen:
		fmt.Println("Door open")
		var pair DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		fmt.Println("Behaviour", (*e).Behaviour)

		switch (*e).Behaviour {
		case EB_DoorOpen:
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
			(*e) = ClearAtCurrentFloor((*e))
			setAllLights(e)

		//lagt inn selv
		case EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
		//

		case EB_Idle:
			fmt.Println("idle")
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)

		}
	}
}

func setAllLights(e *Elevator) {
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
