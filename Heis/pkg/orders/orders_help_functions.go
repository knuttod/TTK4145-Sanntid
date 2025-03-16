package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	// "Heis/pkg/deepcopy"
	"fmt"

	// "fmt"

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

	// fmt.Println("active", activeElevators)


	for _, id := range activeElevators {
		for floor := range N_floors {
			for btn := range N_buttons {
				currentState = (*AssignedOrders)[id][floor][btn]
				updateState = Elevators[remoteId].AssignedOrders[id][floor][btn]

				switch currentState {
				case elevator.Ordr_None:
					switch updateState {
					case elevator.Ordr_Unconfirmed:
						setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Unconfirmed)
					case elevator.Ordr_Confirmed:
						setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Confirmed)
					//Not this case to prevent un resetting cyclic counter
					// case elevator.Ordr_Complete:
					// 	setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Unconfirmed:
					switch updateState {
					case elevator.Ordr_Confirmed:
						setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Confirmed)
					case elevator.Ordr_Complete:
						setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Confirmed:
					switch updateState {
					case elevator.Ordr_Complete:
						setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Complete:
					//top of cyclic counter, reset if synced or already reset
					switch updateState {
					case elevator.Ordr_None:
						setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_None)
					case elevator.Ordr_Complete:
						clearOrder(AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn)
					}

				case elevator.Ordr_Unknown:
					//set to same state as remote
					if updateState != elevator.Ordr_Unknown {
						setOrder(AssignedOrders, id, floor, btn, updateState)
					}
				}

				// switch updateState {
				// case elevator.Ordr_None:
				// 	switch currentState {
				// 	case elevator.Ordr_Unknown:
				// 		setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_None)
				// 	}

				// case elevator.Ordr_Unconfirmed:
				// 	switch currentState {
				// 	case elevator.Ordr_Unknown:
				// 		setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Unconfirmed)
				// 	case elevator.Ordr_None:
				// 		setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Unconfirmed)
				// 	}

				// case elevator.Ordr_Confirmed:
				// 	switch currentState {
				// 	case elevator.Ordr_Unknown:
				// 		setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Confirmed)
				// 	case elevator.Ordr_None:
				// 		setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Confirmed)
				// 	case elevator.Ordr_Unconfirmed:
				// 		setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Confirmed)
				// 	}

				// case elevator.Ordr_Complete:
				// 	// switch currentState {
				// 	// case elevator.Ordr_Unknown:
				// 	// 	setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
				// 	// case elevator.Ordr_None:
				// 	// 	setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
				// 	// case elevator.Ordr_Unconfirmed:
				// 	// 	setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
				// 	// case elevator.Ordr_Confirmed:
				// 	// 	setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
				// 	// case elevator.Ordr_Complete:
				// 	// 	//both top of cyclic counter, check for reset
				// 	// }
				// 	setOrder(AssignedOrders, id, floor, btn, elevator.Ordr_Complete)
				// 	//check if all nodes are at top, in that case reset
				// 	clearOrder(AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn)


				// }
			}
		}
	}
}

// Check if all nodes on nettwork are synced on certain order
func ordersSynced(AssignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) bool {
	if len(activeElevators) == 1 {
		return true
	}

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

// Should start order when elevators are synced and order is unconfirmed or confirmed and no order in elevators map
func shouldStartLocalOrder(AssignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, floor, btn int) bool {
	return ordersSynced(AssignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn) && (((AssignedOrders)[selfId][floor][btn] == elevator.Ordr_Unconfirmed) || (((AssignedOrders)[selfId][floor][btn] == elevator.Ordr_Confirmed) && !Elevators[selfId].Elevator.LocalOrders[floor][btn]))
}

// If FSM is ready, set order to confirmed and starts order if it should start
func confirmAndStartOrder(AssignedOrders *map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int,
	localAssignedOrder chan elevio.ButtonEvent) {
	// if ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn) && (((*AssignedOrders)[id][floor][btn] == elevator.Ordr_Unconfirmed) || (((*AssignedOrders)[id][floor][btn] == elevator.Ordr_Confirmed) && !Elevators[selfId].Elevator.LocalOrders[floor][btn])) {
	if shouldStartLocalOrder(*AssignedOrders, Elevators, activeElevators, selfId, floor, btn) {	
		fmt.Println("start")
		order := elevio.ButtonEvent{
			Floor:  floor,
			Button: elevio.ButtonType(btn),
		}
		select {
		case localAssignedOrder <- order:
			setOrder(AssignedOrders, selfId, floor, btn, elevator.Ordr_Confirmed)
		default:
				//non blocking
		}
		
		
	}
}

func clearOrder(AssignedOrders *map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) {
	if ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn) && ((*AssignedOrders)[id][floor][btn] == elevator.Ordr_Complete) {
		setOrder(AssignedOrders, selfId, floor, btn, elevator.Ordr_None)
	}
}



// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet.
func updateFromRemoteElevator(AssignedOrders *map[string][][]elevator.OrderState, Elevators *map[string]elevator.NetworkElevator, remoteElevatorState msgTypes.ElevatorStateMsg) {
	remoteElevator := remoteElevatorState.NetworkElevator
	(*Elevators)[remoteElevatorState.Id] = remoteElevator

	_, exists := (*AssignedOrders)[remoteElevatorState.Id]
	if !exists {
		(*AssignedOrders)[remoteElevatorState.Id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
	}
}

// Checks if all active elevators in Elevators have an assignedOrders map with keys for all active elevators on nettwork
func assignedOrdersKeysCheck(Elevators map[string]elevator.NetworkElevator, activeElevators []string) bool {

	var assignedOrdersKeys []string
	sort.Strings(activeElevators)

	for _, elev := range Elevators {
		if len(activeElevators) > len(elev.AssignedOrders) {
			return false
		}
		assignedOrdersKeys = []string{}
		for k, _ := range elev.AssignedOrders {
			assignedOrdersKeys = append(assignedOrdersKeys, k)
		}
		sort.Strings(assignedOrdersKeys)

		if !reflect.DeepEqual(activeElevators, assignedOrdersKeys[:len(activeElevators)]) {
			return false
		}
	}
	return true
}

// Variant with referance
func setOrder(orderMap *map[string][][]elevator.OrderState, elevId string, floor, btn int, state elevator.OrderState) {
	temp := (*orderMap)[elevId]
	temp[floor][btn] = state
	(*orderMap)[elevId] = temp
}

//Variant with shallow copy
// func setOrder(assignedOrders [][]elevator.OrderState, elevId string, floor, btn int, state elevator.OrderState) [][]elevator.OrderState{
// 	assignedOrders[floor][btn] = state
// 	return assignedOrders
// }

//Variant with deepCopy
// func setOrder(assignedOrders [][]elevator.OrderState, elevId string, floor, btn int, state elevator.OrderState) [][]elevator.OrderState{
// 	copy := deepcopy.DeepCopy2DSlice(assignedOrders)
// 	copy[floor][btn] = state
// 	return copy
// }


