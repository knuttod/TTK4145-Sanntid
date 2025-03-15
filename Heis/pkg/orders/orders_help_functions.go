package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	// "fmt"
	"reflect"
	"sort"
)

func AssignedOrdersInit(id string) map[string][][]elevator.RequestState {
	assignedOrders := map[string][][]elevator.RequestState{}
	
	Orders := make([][]elevator.RequestState, N_floors)
	for i := range Orders {
		Orders[i] = make([]elevator.RequestState, N_buttons)
	}
	assignedOrders[id] = Orders

	return assignedOrders
}

// Updates state in assignedOrders for local elevator if it is behind remoteElevators
func orderMerger(AssignedOrders *map[string][][]elevator.RequestState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string) {
	var currentState elevator.RequestState
	var updateState elevator.RequestState

	for id, _ := range (*AssignedOrders) {	
		for floor := range N_floors{
			for btn := range N_buttons {
				// for _, elev := range Elevators {
				// 	currentState = (*AssignedOrders)[id][floor][btn]
				// 	updateState = elev.AssignedOrders[id][floor][btn]

				// 	if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
				// 		temp := (*AssignedOrders)[id]
				// 		temp[floor][btn] = updateState
				// 		(*AssignedOrders)[id] = temp
				// 	}
				// }
				for _, elev := range activeElevators {
					if elev == selfId {
						continue
					}
					currentState = (*AssignedOrders)[id][floor][btn]
					updateState = Elevators[elev].AssignedOrders[id][floor][btn]

					if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
						temp := (*AssignedOrders)[id]
						temp[floor][btn] = updateState
						(*AssignedOrders)[id] = temp
					}
				}
				confirmOrCloseOrders(AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn)
			}
		}
	}
}

// Check if all nodes on nettwork are synced on certain order
func ordersSynced(AssignedOrders map[string][][]elevator.RequestState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) bool {
	// for _, elev := range remoteElevators {
	// 	if (*e).AssignedOrders[id][floor][btn] != elev.AssignedOrders[id][floor][btn] {
	// 		return false
	// 	}
	// }
	// return true
	// if len(activeElevators) == 1 {
	// 	return true
	// }
	for _, elev := range activeElevators {
		if elev == selfId {
			continue
		}
		if AssignedOrders[id][floor][btn] != Elevators[elev].AssignedOrders[id][floor][btn] {
			return false
		}
	}
	return true
}


// Increments cyclic counter if all nodes on nettwork are synced on order and state is order or completed
func confirmOrCloseOrders(AssignedOrders *map[string][][]elevator.RequestState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) {
	if ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn) {
		if  (*AssignedOrders)[id][floor][btn] == elevator.Order {
			temp := (*AssignedOrders)[id]
			temp[floor][btn] = elevator.Confirmed
			(*AssignedOrders)[id] = temp
		}
		if  (*AssignedOrders)[id][floor][btn] == elevator.Complete{
			temp := (*AssignedOrders)[id]
			temp[floor][btn] = elevator.None
			(*AssignedOrders)[id] = temp
		}
	}
}

// Should start order when elevators are synced and order is confirmed
func shouldStartLocalOrder(AssignedOrders map[string][][]elevator.RequestState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, floor, btn int) bool {
	return ordersSynced(AssignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn) && AssignedOrders[selfId][floor][btn] == elevator.Confirmed
}

// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet. 
func updateFromRemoteElevator(AssignedOrders * map[string][][]elevator.RequestState, Elevators * map[string]elevator.NetworkElevator, remoteElevatorState msgTypes.ElevatorStateMsg) {
	remoteElevator := remoteElevatorState.NetworkElevator
	(*Elevators)[remoteElevatorState.Id] = remoteElevator
	
	for id, _ := range *Elevators {
		_, exists := (*AssignedOrders)[id]
		if !exists {
			(*AssignedOrders)[id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
		}
	}
}

// Checks if all active elevators in remoteElevators have an assignedOrders map with keys for all active elevators on nettwork 
func assignedOrdersKeysCheck(AssignedOrders map[string][][]elevator.RequestState, Elevators map[string]elevator.NetworkElevator, selfId string, activeElevators []string) bool {
	
	var localKeys []string
	var remoteKeys []string
	
	for k, _ := range AssignedOrders{
		localKeys = append(localKeys, k)
	}

	// for _, elev := range remoteElevators {
	// 	remoteKeys = []string{}
	// 	for k, _ := range elev.AssignedOrders{
	// 		remoteKeys = append(remoteKeys, k)
	// 	}
	// 	sort.Strings(localKeys)
	// 	sort.Strings(remoteKeys)
	// 	// fmt.Println("local keys", localKeys)
	// 	// fmt.Println("External keys", remoteKeys)

	// 	if !reflect.DeepEqual(localKeys, remoteKeys) {
	// 		return false
	// 	}
	// }
	for _, elev := range activeElevators {
		if elev == selfId {
			continue
		}
		remoteKeys = []string{}
		for k, _ := range Elevators[elev].AssignedOrders{
			remoteKeys = append(remoteKeys, k)
		}
		sort.Strings(localKeys)
		sort.Strings(remoteKeys)
		// fmt.Println("local keys", localKeys)
		// fmt.Println("External keys", remoteKeys)

		if !reflect.DeepEqual(localKeys, remoteKeys) {
			return false
		}
	}
	return true
}

func restartOrdersSynchroniser(AssignedOrders * map[string][][]elevator.RequestState, remoteElevatorState msgTypes.ElevatorStateMsg) {

	var update bool = false
	for id, _ := range *AssignedOrders {
		for floor := range N_floors {
			for btn := range N_buttons {
				if remoteElevatorState.NetworkElevator.AssignedOrders[id][floor][btn] != elevator.None {
					update = true
					break;
				}
			}
		}
		if update {
			(*AssignedOrders)[id] = remoteElevatorState.NetworkElevator.AssignedOrders[id]
		}
	}
}

func reconnectOrdersSynchroniser(AssignedOrders * map[string][][]elevator.RequestState, remoteElevatorState msgTypes.ElevatorStateMsg, activeElevators []string) {
	var stateLocal elevator.RequestState
	var stateRemote elevator.RequestState

	for id, _ := range *AssignedOrders {
		for floor := range N_floors {
			for btn := range N_buttons {
				stateLocal = (*AssignedOrders)[id][floor][btn]
				stateRemote = remoteElevatorState.NetworkElevator.AssignedOrders[id][floor][btn]

				if stateLocal < stateRemote {
					temp := (*AssignedOrders)[id]
					temp[floor][btn] = stateRemote
					(*AssignedOrders)[id] = temp
				} else {
					temp := remoteElevatorState.NetworkElevator.AssignedOrders[id]
					temp[floor][btn] = stateLocal
					remoteElevatorState.NetworkElevator.AssignedOrders[id] = temp
				}
			}
		}
	}

}

func setAllHallLightsfromRemote(Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string) {
	
	var setLight bool
	for floor := 0; floor < N_floors; floor++ {
		for btn := 0; btn < N_buttons; btn++ {
			setLight = false
			for _, elev := range activeElevators {
				if elev == selfId {
					continue
				}
				if Elevators[elev].AssignedOrders[elev][floor][btn] == elevator.Confirmed{
					setLight = true	
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, setLight)
		}
	}
}