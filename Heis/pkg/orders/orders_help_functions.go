package orders

import (
	"Heis/pkg/deepcopy"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/network/message"
	"Heis/pkg/network/peers"
	// "fmt"
	// "reflect"
	// "sort"
)

func assignedOrdersInit(id string) map[string][][]elevator.OrderState {
	assignedOrders := map[string][][]elevator.OrderState{}

	Orders := make([][]elevator.OrderState, numFloors)
	for i := range numFloors {
		Orders[i] = make([]elevator.OrderState, numBtns)
		for j := range numBtns {
			Orders[i][j] = elevator.Ordr_Unknown
		}
	}
	assignedOrders[id] = Orders

	return assignedOrders
}

// Updates state in assignedOrders for local elevator if it is behind remoteElevators
func orderMerger(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, remoteId string) map[string][][]elevator.OrderState {
	var currentState elevator.OrderState
	var updateState elevator.OrderState

	for _, elevId := range activeElevators {
		for floor := range numFloors {
			for btn := range numBtns {
				currentState = assignedOrders[elevId][floor][btn]
				updateState = elevators[remoteId].AssignedOrders[elevId][floor][btn]

				switch currentState {
				case elevator.Ordr_None:
					//bottom of cyclic counter
					switch updateState {
					case elevator.Ordr_Unconfirmed:
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Unconfirmed)
					case elevator.Ordr_Confirmed:
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Confirmed)
					case elevator.Ordr_Complete:
						// do nothing to prevent un-resetting cyclic counter
					}

				case elevator.Ordr_Unconfirmed:
					switch updateState {
					case elevator.Ordr_Confirmed:
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Confirmed)
					case elevator.Ordr_Complete:
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Confirmed:
					switch updateState {
					case elevator.Ordr_Complete:
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Complete:
					//top of cyclic counter, reset if synced or already reset
					switch updateState {
					case elevator.Ordr_None:
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_None)
					case elevator.Ordr_Complete:
						//resets cyclic counter if all elevators have complete
						if ordersSynced(assignedOrders, elevators, activeElevators, elevId, floor, btn)  {
							assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_None)
						}
					}

				case elevator.Ordr_Unknown:
					if updateState != elevator.Ordr_Unknown {
						//set to same state as remote
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, updateState)
					}
				}
			}
		}
	}
	return assignedOrders
}

// Check if all nodes on nettwork are synced on certain order
func ordersSynced(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, elevId string, floor, btn int) bool {
	if len(activeElevators) == 1 {
		return true
	}

	for _, elev := range activeElevators {
		if elev == selfId {
			continue
		}
		if assignedOrders[elevId][floor][btn] != elevators[elev].AssignedOrders[elevId][floor][btn] {
			return false
		}
	}
	return true
}

// Should start order when elevators are synced and order is unconfirmed or confirmed and no order in elevators map
func shouldStartLocalOrder(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, floor, btn int) bool {
	return ordersSynced(assignedOrders, elevators, activeElevators, selfId, floor, btn) && (((assignedOrders)[selfId][floor][btn] == elevator.Ordr_Unconfirmed) || (((assignedOrders)[selfId][floor][btn] == elevator.Ordr_Confirmed) && !elevators[selfId].Elevator.LocalOrders[floor][btn]))
}

// If FSM is ready, set order to confirmed and starts order if it should start
func confirmAndStartOrder(assignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, elevId string, floor, btn int,
	localAssignedOrder chan elevio.ButtonEvent) map[string][][]elevator.OrderState{

	if shouldStartLocalOrder(assignedOrders, Elevators, activeElevators, floor, btn) {
		order := elevio.ButtonEvent{
			Floor:  floor,
			Button: elevio.ButtonType(btn),
		}
		//sends order to fsm if it is ready to take it
		select {
		case localAssignedOrder <- order:
			assignedOrders[selfId] = setOrder(assignedOrders[selfId], floor, btn, elevator.Ordr_Confirmed)
		default:
		}
	}
	return assignedOrders
}

// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet.
func updateFromRemoteElevator(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, remoteElevatorState message.ElevatorStateMsg) (map[string][][]elevator.OrderState, map[string]elevator.NetworkElevator) {
	remoteElevator := remoteElevatorState.NetworkElevator
	elevators[remoteElevatorState.Id] = remoteElevator

	_, exists := assignedOrders[remoteElevatorState.Id]
	if !exists {
		assignedOrders[remoteElevatorState.Id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
	}

	return assignedOrders, elevators
}

// Needs this check to prevent index error. Checks if all active elevators in Elevators have an assignedOrders map with keys for all active elevators on nettwork. 
func assignedOrdersKeysCheck(elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string) bool {

	if len(activeElevators) == 0 {
		return false
	}

	if len(activeElevators) == 1 {
		if activeElevators[0] == selfId {
			return true
		}
		return false
	}

	var assignedOrdersKeys map[string]bool
	for _, id := range activeElevators {
		elev := elevators[id]
		if len(activeElevators) > len(elev.AssignedOrders) {
			return false
		}
		assignedOrdersKeys = make(map[string]bool)
		for id, _ := range elev.AssignedOrders {
			assignedOrdersKeys[id] = true
		}

		for _, elev := range activeElevators {
			if !assignedOrdersKeys[elev] {
				return false
			}
		}
	}
	return true
}


//Variant with shallow copy
func setOrder(orders [][]elevator.OrderState, floor, btn int, state elevator.OrderState) [][]elevator.OrderState{
	orders[floor][btn] = state
	return orders
}

//Variant with deepCopy
// func setOrder(orders [][]elevator.OrderState, elevId string, floor, btn int, state elevator.OrderState) [][]elevator.OrderState{
// 	copy := deepcopy.DeepCopy2DSlice(orders)
// 	copy[floor][btn] = state
// 	return copy
// }


func peerUpdateHandler(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, p peers.PeerUpdate) map[string][][]elevator.OrderState{

	//detects disconnected elevators and reassigns their orders
	if len(p.Lost) > 0 {

		if assignedOrdersKeysCheck(elevators, activeElevators, selfId) {
			reassignOrdersFromDisconnectedElevators(assignedOrders, deepcopy.DeepCopyElevatorsMap(elevators),  p.Lost, activeElevators, selfId)
		}
	}

	//if elevator(s) has only disconnected and reconnects (not crashed) it has its latest info about its order and to not override it having no order the order is set to complete
	if len(p.New) > 0 {
		for floor := range numFloors {
			if (assignedOrders)[selfId][floor][int(elevio.BT_Cab)] == elevator.Ordr_None {
				assignedOrders[selfId] = setOrder(assignedOrders[selfId], floor, int(elevio.BT_Cab), elevator.Ordr_Complete)
			}
		}
	}

	//fikse clearing av hall orders etter tilkoblinkg på nettet

	//tror denne gjør akkurat det samme som den to hakk over
	//sets orders on all other elevators to unkwown, since information can not be trusted
	if len(p.Peers) == 1 {
		for id := range assignedOrders {
			// Kanskje ikke sette sine egne til unkown
			if id == selfId {
				continue
			}
			for floor := range numFloors {
				for btn := range numBtns - 1 {
					assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Unknown)
				}
			}
		}
	}
	return assignedOrders
}

func setHallLights(assignedOrders map[string][][]elevator.OrderState, activeElevators []string, activeHallLights [][]bool) [][]bool {
	for floor := range numFloors {
		for btn := range numBtns - 1 {
			setLight := false
			for _, elev := range activeElevators {
				if assignedOrders[elev][floor][btn] == elevator.Ordr_Confirmed {
					setLight = true
					break
				}
			}
			if setLight != activeHallLights[floor][btn] {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, setLight)
				activeHallLights[floor][btn] = setLight
			}
		}
	}
	return activeHallLights
}

func initHallLights() [][]bool {
	activeHallLights := make([][]bool, numFloors)
	for floor := range numFloors {
		activeHallLights[floor] = make([]bool, numBtns-1)
		for btn := range numBtns - 1 {
			activeHallLights[floor][btn] = false
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
		}
	}
	return activeHallLights
}
