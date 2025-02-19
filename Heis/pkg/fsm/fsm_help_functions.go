package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/elevator"
	//"fmt"
)

// func fsm_init(e *elevator.Elevator) {
// 	// initialize the (*e) struct
// 	(*e).Floor = -1
// 	(*e).Dirn = elevio.MD_Stop
// 	(*e).Behaviour = elevator.EB_Idle
// 	(*e).Config.ClearRequestVariant = elevator.CV_InDirn
// 	(*e).Config.DoorOpenDuration_s = 3.0
// 	(*e).Requests = make([][]bool, N_floors)
// 	for i := range (*e).Requests {
// 		(*e).Requests[i] = make([]bool, N_buttons)
// 	}

// }

func initBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = elevator.EB_Moving
}

func requestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64) {
	//print functions??
	// buttonPressMsg := msgTypes.ButtonPressMsg{
	// 	Floor:  btn_floor,
	// 	Button: btn_type,
	// 	Id:     id,
	// }

	// // Retransmit to reduce redundancy
	// for i := 0; i < 30; i++ {
	// 	Tx <- msgTypes.UdpMsg{ButtonPressMsg: &buttonPressMsg}
	// 	time.Sleep(10 * time.Millisecond)
	// }

	// elevatorStateMsg := msgTypes.ElevatorStateMsg{
	// 	Elevator: &e,
	// 	Id:       id,
	// }

	// // Retransmit to reduce redundancy
	// for i := 0; i < 30; i++ {
	// 	Tx <- msgTypes.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
	// 	time.Sleep(10 * time.Millisecond)
	// }



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
