package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"reflect"
	"sort"
)

// Updates state in assignedOrders for local elevator if it is behind remoteElevators
func orderMerger(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string) {
	var currentState elevator.RequestState
	var updateState elevator.RequestState

	for id, _ := range (*e).AssignedOrders {	
		for floor := range N_floors{
			for btn := range N_buttons {
				// for _, elev := range remoteElevators {
				// 	currentState = (*e).AssignedOrders[id][floor][btn]
				// 	updateState = elev.AssignedOrders[id][floor][btn]

				// 	if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
				// 		temp := (*e).AssignedOrders[id]
				// 		temp[floor][btn] = updateState
				// 		(*e).AssignedOrders[id] = temp
				// 	}
				// }
				for _, elev := range activeElevators {
					if elev == (*e).Id {
						continue
					}
					currentState = (*e).AssignedOrders[id][floor][btn]
					updateState = remoteElevators[elev].AssignedOrders[id][floor][btn]

					if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
						temp := (*e).AssignedOrders[id]
						temp[floor][btn] = updateState
						(*e).AssignedOrders[id] = temp
					}
				}	
				confirmOrCloseOrders(e, remoteElevators, activeElevators, id, floor, btn)
			}
		}
	}
}

// Check if all nodes on nettwork are synced on certain order
func ordersSynced(e elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, id string, floor, btn int) bool {
	// for _, elev := range remoteElevators {
	// 	if (*e).AssignedOrders[id][floor][btn] != elev.AssignedOrders[id][floor][btn] {
	// 		return false
	// 	}
	// }
	// return true
	for _, elev := range activeElevators {
		if elev == (e).Id {
			continue
		}
		if (e).AssignedOrders[id][floor][btn] != remoteElevators[elev].AssignedOrders[id][floor][btn] {
			return false
		}
	}
	return true
}


// Increments cyclic counter if all nodes on nettwork are synced on order and state is order or completed
func confirmOrCloseOrders(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, id string, floor, btn int) {
	if ordersSynced(*e, remoteElevators, activeElevators, id, floor, btn) {
		if  (*e).AssignedOrders[id][floor][btn] == elevator.Order {
			temp := (*e).AssignedOrders[id]
			temp[floor][btn] = elevator.Confirmed
			(*e).AssignedOrders[id] = temp
		}
		if  (*e).AssignedOrders[id][floor][btn] == elevator.Complete{
			temp := (*e).AssignedOrders[id]
			temp[floor][btn] = elevator.None
			(*e).AssignedOrders[id] = temp
		}
	}
}

// Should start order when elevators are synced and order is confirmed
func shouldStartLocalOrder(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, id string, floor, btn int) bool {
	return ordersSynced(*e, remoteElevators, activeElevators, id, floor, btn) && (*e).AssignedOrders[id][floor][btn] == elevator.Confirmed
}

// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet. 
func updateFromRemoteElevator(remoteElevators * map[string]elevator.Elevator, e * elevator.Elevator, remoteElevatorState msgTypes.ElevatorStateMsg) {
	remoteElevator := remoteElevatorState.Elevator
	(*remoteElevators)[remoteElevatorState.Id] = remoteElevator
	
	for id, _ := range *remoteElevators {
		_, exists := (*e).AssignedOrders[id]
		if !exists {
			(*e).AssignedOrders[id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
		}
	}
}

// Checks if all active elevators in remoteElevators have an assignedOrders map with keys for all active elevators on nettwork 
func assignedOrdersKeysCheck(remoteElevators map[string]elevator.Elevator, e elevator.Elevator, activeElevators []string) bool {
	
	var localKeys []string
	var remoteKeys []string
	
	for k, _ := range e.AssignedOrders{
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
		if elev == e.Id {
			continue
		}
		remoteKeys = []string{}
		for k, _ := range remoteElevators[elev].AssignedOrders{
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

func restartOrdersSynchroniser(e * elevator.Elevator, remoteElevatorState msgTypes.ElevatorStateMsg) {

	var update bool = false
	for id, _ := range (*e).AssignedOrders {
		for floor := range N_floors {
			for btn := range N_buttons {
				if remoteElevatorState.Elevator.AssignedOrders[id][floor][btn] != elevator.None {
					update = true
					break;
				}
			}
		}
		if update {
			(*e).AssignedOrders[id] = remoteElevatorState.Elevator.AssignedOrders[id]
		}
	}
}

func reconnectOrdersSynchroniser(e * elevator.Elevator, remoteElevatorState msgTypes.ElevatorStateMsg, activeElevators []string) {
	var stateLocal elevator.RequestState
	var stateRemote elevator.RequestState

	for id, _ := range (*e).AssignedOrders {
		for floor := range N_floors {
			for btn := range N_buttons {
				stateLocal = (*e).AssignedOrders[id][floor][btn]
				stateRemote = remoteElevatorState.Elevator.AssignedOrders[id][floor][btn]

				if stateLocal < stateRemote {
					temp := (*e).AssignedOrders[id]
					temp[floor][btn] = stateRemote
					(*e).AssignedOrders[id] = temp
				} else {
					temp := remoteElevatorState.Elevator.AssignedOrders[id]
					temp[floor][btn] = stateLocal
					remoteElevatorState.Elevator.AssignedOrders[id] = temp
				}
			}
		}
	}

}

func setAllHallLightsfromRemote(remoteElevators map[string]elevator.Elevator, activeElevators []string, selfId string) {
	
	var setLight bool
	for floor := 0; floor < N_floors; floor++ {
		for btn := 0; btn < N_buttons; btn++ {
			setLight = false
			for _, elev := range activeElevators {
				if elev == selfId {
					continue
				}
				if remoteElevators[elev].AssignedOrders[elev][floor][btn] == elevator.Confirmed{
					setLight = true	
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, setLight)
		}
	}
}