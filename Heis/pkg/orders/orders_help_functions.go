package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	// "fmt"
	"reflect"
	"sort"
)

func AssignedOrdersInit(id string) map[string][][]elevator.OrderState {
	assignedOrders := map[string][][]elevator.OrderState{}
	
	Orders := make([][]elevator.OrderState, N_floors)
	for i := range N_floors {
		Orders[i] = make([]elevator.OrderState, N_buttons)
		for j := range N_buttons {
			Orders[i][j] = elevator.Ordr_Unknown
		}
	}
	assignedOrders[id] = Orders

	return assignedOrders
}

// Updates state in assignedOrders for local elevator if it is behind remoteElevators
func orderMerger(AssignedOrders *map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, remoteId string) {
	var currentState elevator.OrderState
	var updateState elevator.OrderState

	for _, id := range activeElevators {	
		for floor := range N_floors{
			for btn := range N_buttons {
				currentState = (*AssignedOrders)[id][floor][btn]
				updateState = Elevators[remoteId].AssignedOrders[id][floor][btn]

				switch currentState{
				case elevator.Ordr_None:
					switch updateState {
					case elevator.Ordr_Unconfirmed:
						setOrder(AssignedOrders, selfId, floor, btn, elevator.Ordr_Unconfirmed)
					case elevator.Ordr_Confirmed:
						setOrder(AssignedOrders, selfId, floor, btn, elevator.Ordr_Confirmed)
					}
				
				case elevator.Ordr_Unconfirmed:
					switch updateState {
					case elevator.Ordr_Confirmed:
						setOrder(AssignedOrders, selfId, floor, btn, elevator.Ordr_Confirmed)
					}
				
				case elevator.Ordr_Confirmed:
					//do nothing, top of cyclic counter

				case elevator.Ordr_Unknown:
					//set to same state as remote
					setOrder(AssignedOrders, selfId, floor, btn, updateState)
				}

				//Confirms order if it should be confirmed
				confirmOrder(AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn)
			}
		}
	}
}

// Check if all nodes on nettwork are synced on certain order
func ordersSynced(AssignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) bool {
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

//Set order to confirmed if it is unconfirmed for all elevators
func confirmOrder(AssignedOrders *map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) {
	if ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn) && (*AssignedOrders)[id][floor][btn] == elevator.Ordr_Unconfirmed {
		setOrder(AssignedOrders, selfId, floor, btn, elevator.Ordr_Confirmed)
	}
}


// Should start order when elevators are synced and order is confirmed
func shouldStartLocalOrder(AssignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, floor, btn int) bool {
	return ordersSynced(AssignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn) && AssignedOrders[selfId][floor][btn] == elevator.Ordr_Confirmed
}


// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet. 
func updateFromRemoteElevator(AssignedOrders * map[string][][]elevator.OrderState, Elevators * map[string]elevator.NetworkElevator, remoteElevatorState msgTypes.ElevatorStateMsg) {
	remoteElevator := remoteElevatorState.NetworkElevator
	(*Elevators)[remoteElevatorState.Id] = remoteElevator
	
	_, exists := (*AssignedOrders)[remoteElevatorState.Id]
	if !exists {
		(*AssignedOrders)[remoteElevatorState.Id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
	}
}

// Checks if all active elevators in remoteElevators have an assignedOrders map with keys for all active elevators on nettwork 
func assignedOrdersKeysCheck(AssignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, selfId string, activeElevators []string) bool {
	
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

func restartOrdersSynchroniser(AssignedOrders * map[string][][]elevator.OrderState, remoteElevatorState msgTypes.ElevatorStateMsg) {

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

func reconnectOrdersSynchroniser(AssignedOrders * map[string][][]elevator.OrderState, remoteElevatorState msgTypes.ElevatorStateMsg, activeElevators []string) {
	var stateLocal elevator.OrderState
	var stateRemote elevator.OrderState

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

func setOrder(orderMap *map[string][][]elevator.OrderState, elevId string, floor, btn int, state elevator.OrderState) {
	temp := (*orderMap)[elevId]
	temp[floor][btn] = state
	(*orderMap)[elevId] = temp
}


// func changeElevatorMap