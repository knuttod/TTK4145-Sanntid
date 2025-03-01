package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"fmt"
	"reflect"
	"sort"

	// "reflect"
	"time"
	//"fmt"
	// "reflect"
)

// func fsm_init(e *elevator.Elevator) {
// 	// initialize the (*e) struct
// 	(*e).Floor = -2
// 	(*e).Dirn = elevio.MD_Stop
// 	(*e).Behaviour = elevator.EB_Idle
// 	(*e).Config.ClearRequestVariant = elevator.CV_InDirn
// 	(*e).Config.DoorOpenDuration_s = 3.0
// 	(*e).AssignedOrders = make([][]bool, N_floors)
// 	for i := range (*e).AssignedOrders {
// 		(*e).AssignedOrders[i] = make([]bool, N_buttons)
// 	}

// }

func initBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = elevator.EB_Moving
}

func requestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64, Tx chan msgTypes.UdpMsg, id string) {

	switch (*e).Behaviour {
	case elevator.EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
			temp := (*e).AssignedOrders[(*e).Id]
			temp[btn_floor][btn_type] = elevator.Complete
			(*e).AssignedOrders[(*e).Id] = temp
			// (*e).AssignedOrders[(*e).Id][btn_floor][btn_type] = elevator.Complete
		} else {
			//(*e).AssignedOrders[btn_floor][btn_type] = elevator.Confirmed
		}

	case elevator.EB_Moving:
		//(*e).AssignedOrders[btn_floor][btn_type] = elevator.Confirmed

	case elevator.EB_Idle:
		//(*e).AssignedOrders[btn_floor][btn_type] = elevator.Confirmed
		var pair elevator.DirnBehaviourPair = ChooseDirection((*e))
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
	// fmt.Printf("Elevator %s Requests before sending: %+v\n", id, e.Requests)

	elevatorStateMsg := msgTypes.ElevatorStateMsg{
		Elevator: *e,
		Id:       id,
	}

	// Retransmit to reduce redundancy
	for i := 0; i < 10; i++ {
		Tx <- msgTypes.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
		time.Sleep(10 * time.Millisecond)
	}

	setAllLights(e)
}

func floorArrival(e *elevator.Elevator, newFloor int, drv_doorTimer chan float64, Tx chan msgTypes.UdpMsg, id string) {
	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)
	fmt.Println("hei")

	switch (*e).Behaviour {
	case elevator.EB_Moving:
		fmt.Println("moving")
		if ShouldStop(*e) {
			fmt.Println("STOP")
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			*e = ClearAtCurrentFloor(*e)
			drv_doorTimer <- e.Config.DoorOpenDuration_s
			setAllLights(e)
			(*e).Behaviour = elevator.EB_DoorOpen

			// **Log current state before sending**
			// fmt.Printf("Elevator %s Requests before sending: %+v\n", id, e.Requests)

			// elevatorStateMsg := msgTypes.ElevatorStateMsg{
			// 	Elevator: *e,
			// 	Id:       id,
			// }

			// // Retransmit to reduce redundancy
			// for i := 0; i < 10; i++ {
			// 	Tx <- msgTypes.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
			// 	time.Sleep(10 * time.Millisecond)
			// }

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
			if btn == 2 {
				if e.AssignedOrders[(*e).Id][floor][btn] == elevator.Confirmed {
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
				} else {
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
				}
			} else {
				buttonLightOn := false
				for id, _ := range (*e).AssignedOrders {
					if e.AssignedOrders[id][floor][btn] == elevator.Confirmed {
						buttonLightOn = true
						break
					}
				}
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, buttonLightOn)
			}
			
		}
	}
}

// func SetAllLightsOrder(Orders [][]int, e *elevator.Elevator) {
// 	//set ligths
// 	for floor := range Orders {
// 		for btn := 0; btn < 2; btn++ {
// 			if Orders[floor][btn] == elevator.Confirmed {
// 				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
// 			} else {
// 				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
// 			}
// 		}
// 		if Orders[floor][(*e).Index+1] == elevator.Confirmed {
// 			elevio.SetButtonLamp(elevio.ButtonType(2), floor, true)
// 		} else {
// 			elevio.SetButtonLamp(elevio.ButtonType(2), floor, false)
// 		}
// 	}
// }

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
func mergeRequests(local *elevator.Elevator, remote elevator.Elevator) {
	fmt.Println("Copying remote requests for debugging:")

	// local.Requests = remote.Requests
	setAllLights(local)

	fmt.Println("Local requests after copying:", *local)
}

// func sendStates(local *elevator.Elevator, remote *elevator.Elevator, stateUpdated chan bool) {
// 	for {
// 		if local != remote {
// 			stateUpdated <- true
// 		}
// 	}
// }

func sendStates(local *elevator.Elevator, remote elevator.Elevator, stateUpdated chan bool) {
	// fmt.Println("local", (*local).AssignedOrders)
	// fmt.Println("remote", remote.AssignedOrders)
	// if reflect.DeepEqual((*local).AssignedOrders, remote.AssignedOrders) {
	// 	stateUpdated <- true
	// }
	// var localKeys []string
	// for k, _ := range (*local).AssignedOrders {
	// 	localKeys = append(localKeys, k)
	// }
	// var remoteKeys []string
	// for k, _ := range remote.AssignedOrders {
	// 	remoteKeys = append(remoteKeys, k)
	// }
	// if reflect.DeepEqual(localKeys, remoteKeys) {
	// 	stateUpdated <- true
	// }
	stateUpdated <- true
	//if len((*local).AssignedOrders) == len(remote.AssignedOrders) {
	//	fmt.Println(len((*local).AssignedOrders))
	//	stateUpdated <- true
	//}
}

func assignedOrdersCheck(remoteElevators map[string]elevator.Elevator, elevator elevator.Elevator) bool {
	var localKeys []string
	var remoteKeys []string
	

	for k, _ := range elevator.AssignedOrders{
		localKeys = append(localKeys, k)
	}
	// if !reflect.DeepEqual(remoteElevatorKeys, assignedOrdersKeys) {
	// 	return false
	// }
	
	for _, elev := range remoteElevators {
		remoteKeys = []string{}
		for k, _ := range elev.AssignedOrders{
			remoteKeys = append(remoteKeys, k)
		}
		sort.Strings(localKeys)
		sort.Strings(remoteKeys)
		fmt.Println("local keys", localKeys)
		fmt.Println("External keys", remoteKeys)

		if !reflect.DeepEqual(localKeys, remoteKeys) {
			return false
		}
	}
	return true
}