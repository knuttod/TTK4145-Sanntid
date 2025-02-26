package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
	"time"
	//"fmt"
)

// func fsm_init(e *elevator.Elevator) {
// 	// initialize the (*e) struct
// 	(*e).Floor = -2
// 	(*e).Dirn = elevio.MD_Stop
// 	(*e).Behaviour = elevator.EB_Idle
// 	(*e).Config.ClearRequestVariant = elevator.CV_InDirn
// 	(*e).Config.DoorOpenDuration_s = 3.0
// 	(*e).LocalOrders = make([][]bool, N_floors)
// 	for i := range (*e).LocalOrders {
// 		(*e).LocalOrders[i] = make([]bool, N_buttons)
// 	}

// }

func initBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = elevator.EB_Moving
}

func requestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
			(*e).LocalOrders[btn_floor][btn_type] = 3
		} else {
			//(*e).LocalOrders[btn_floor][btn_type] = 2
		}

	case elevator.EB_Moving:
		//(*e).LocalOrders[btn_floor][btn_type] = 2

	case elevator.EB_Idle:
		//(*e).LocalOrders[btn_floor][btn_type] = 2
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

func DoorTimeout(e *elevator.Elevator, drv_doorTimer chan float64) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		var pair elevator.DirnBehaviourPair = ChooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch (*e).Behaviour {
		case elevator.EB_DoorOpen:
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s //????
			//drv_doorTimer <- 0.0
			(*e) = ClearAtCurrentFloor((*e))
			// setAllLights(e)
			SetAllLightsOrder((*e).GlobalOrders, e)

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
			if e.LocalOrders[floor][btn] == 2 {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}

func SetAllLightsOrder(Orders [][]int, e *elevator.Elevator) {
	//set ligths
	for floor := range Orders {
		for btn := 0; btn < 2; btn++ {
			if Orders[floor][btn] == 2 {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
		if Orders[floor][(*e).Index+1] == 2 {
			elevio.SetButtonLamp(elevio.ButtonType(2), floor, true)
		} else {
			elevio.SetButtonLamp(elevio.ButtonType(2), floor, false)
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
func mergeRequests(local *[][]bool, remote [][]bool) {
	fmt.Println("Copying remote requests for debugging:")

	// Copy all requests from remote to local
	for floor := 0; floor < len(*local); floor++ {
		for btn := 0; btn < len((*local)[floor]); btn++ {
			(*local)[floor][btn] = remote[floor][btn] // Directly copying instead of merging
			if remote[floor][btn] {
				fmt.Printf("Copied request at floor %d, button %d\n", floor, btn)
			}
		}
	}

	fmt.Println("Local requests after copying:", *local)
}
