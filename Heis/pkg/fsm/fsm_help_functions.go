package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/types"

	"fmt"
	"time"
)

func fsm_init(e *types.Elevator) {
	// initialize the (*e) struct
	(*e).Floor = -1
	(*e).Dirn = elevio.MD_Stop
	(*e).Behaviour = types.EB_Idle
	(*e).Config.ClearRequestVariant = types.CV_InDirn
	(*e).Config.DoorOpenDuration_s = 3.0
	(*e).Requests = make([][]bool, N_floors)
	for i := range (*e).Requests {
		(*e).Requests[i] = make([]bool, N_buttons)
	}
}

func initBetweenFloors(e *types.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = types.EB_Moving
}

func requestButtonPress(e *types.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64, Tx chan types.UdpMsg, id string) {
	//print functions??

	switch (*e).Behaviour {
	case types.EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
		} else {
			(*e).Requests[btn_floor][btn_type] = true
		}

	case types.EB_Moving:
		(*e).Requests[btn_floor][btn_type] = true

	case types.EB_Idle:
		(*e).Requests[btn_floor][btn_type] = true
		var pair DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case types.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			//drv_doorTimer <- 0.0
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			(*e) = ClearAtCurrentFloor((*e))

		case types.EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
			//clear something at this floor??

		case types.EB_Idle:
			//need something here?
		}

	}

	// **Log current state before sending**
	fmt.Printf("Elevator %s Requests before sending: %+v\n", id, e.Requests)

	elevatorStateMsg := types.ElevatorStateMsg{
		Elevator: *e,
		Id:       id,
	}

	// Retransmit to reduce redundancy
	for i := 0; i < 10; i++ {
		Tx <- types.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
		time.Sleep(10 * time.Millisecond)
	}

	setAllLights(e)
}

func floorArrival(e *types.Elevator, newFloor int, drv_doorTimer chan float64, Tx chan types.UdpMsg, id string) {
	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)

	switch e.Behaviour {
	case types.EB_Moving:
		if ShouldStop(*e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			*e = ClearAtCurrentFloor(*e)
			drv_doorTimer <- e.Config.DoorOpenDuration_s
			setAllLights(e)
			(*e).Behaviour = types.EB_DoorOpen

			// **Log current state before sending**
			fmt.Printf("Elevator %s Requests before sending: %+v\n", id, e.Requests)

			elevatorStateMsg := types.ElevatorStateMsg{
				Elevator: *e,
				Id:       id,
			}

			// Retransmit to reduce redundancy
			for i := 0; i < 10; i++ {
				Tx <- types.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
				time.Sleep(10 * time.Millisecond)
			}

		}
	}
}

func DoorTimeout(e *types.Elevator, drv_doorTimer chan float64) {

	switch (*e).Behaviour {
	case types.EB_DoorOpen:
		var pair DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch (*e).Behaviour {
		case types.EB_DoorOpen:
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s //????
			//drv_doorTimer <- 0.0
			(*e) = ClearAtCurrentFloor((*e))
			setAllLights(e)

		//lagt inn selv
		case types.EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
		//

		case types.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)

		}
	}
}

func setAllLights(e *types.Elevator) {
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

// func mergeRequests(local *[][]bool, remote [][]bool) {
// 	for floor := 0; floor < len(*local); floor++ {
// 		for btn := 0; btn < len((*local)[floor]); btn++ {
// 			// Merge requests: if either elevator has requested this floor/button, keep it.
// 			(*local)[floor][btn] = (*local)[floor][btn] || remote[floor][btn]
// 		}
// 	}
// }

// func mergeRequests(local *[][]bool, remote [][]bool) {
// 	fmt.Println("Merging requests:")
// 	for floor := 0; floor < len(*local); floor++ {
// 		for btn := 0; btn < len((*local)[floor]); btn++ {
// 			oldValue := (*local)[floor][btn]
// 			(*local)[floor][btn] = (*local)[floor][btn] || remote[floor][btn]
// 			if (*local)[floor][btn] != oldValue {
// 				fmt.Printf("Merged request at floor %d, button %d\n", floor, btn)
// 			}
// 		}
// 	}
// }

// DEBUG VERSION Merge does not work but the sending of the struct works atm the elevator does not react to its states changing, maby need a queue of requests handled by a goroutine?
func mergeRequests(local *types.Elevator, remote types.Elevator) {
	fmt.Println("Copying remote requests for debugging:")

	local.Requests = remote.Requests
	setAllLights(local)

	fmt.Println("Local requests after copying:", *local)
}

func sendStates(local *types.Elevator, remote *types.Elevator, stateUpdated chan bool) {
	for {
		if local != remote {
			stateUpdated <- true
		}
	}
}
