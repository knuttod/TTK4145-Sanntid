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

func AssignedOrdersInit(id string) map[string][][]elevator.OrderState {
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
func orderMerger(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, remoteId string) map[string][][]elevator.OrderState {
	var currentState elevator.OrderState
	var updateState elevator.OrderState

	for _, id := range activeElevators {
		for floor := range numFloors {
			for btn := range numBtns {
				currentState = assignedOrders[id][floor][btn]
				updateState = elevators[remoteId].AssignedOrders[id][floor][btn]

				switch currentState {
				case elevator.Ordr_None:
					//bottom of cyclic counter
					switch updateState {
					case elevator.Ordr_Unconfirmed:
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Unconfirmed)
					case elevator.Ordr_Confirmed:
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Confirmed)
						//Not this case to prevent un resetting cyclic counter
						// case elevator.Ordr_Complete:
						// 	assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Unconfirmed:
					switch updateState {
					case elevator.Ordr_Confirmed:
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Confirmed)
					case elevator.Ordr_Complete:
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Confirmed:
					switch updateState {
					case elevator.Ordr_Complete:
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_Complete)
					}

				case elevator.Ordr_Complete:
					//top of cyclic counter, reset if synced or already reset
					switch updateState {
					case elevator.Ordr_None:
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_None)
					case elevator.Ordr_Complete:
						//resets cyclic counter if all elevators have complete
						if ordersSynced(assignedOrders, elevators, activeElevators, selfId, id, floor, btn)  {
							assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_None)
						}
					}

				case elevator.Ordr_Unknown:
					//set to same state as remote
					if updateState != elevator.Ordr_Unknown {
						assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, updateState)
					}
				}
			}
		}
	}
	return assignedOrders
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
func confirmAndStartOrder(assignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int,
	localAssignedOrder chan elevio.ButtonEvent) map[string][][]elevator.OrderState{
	// if ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, id, floor, btn) && (((*AssignedOrders)[id][floor][btn] == elevator.Ordr_Unconfirmed) || (((*AssignedOrders)[id][floor][btn] == elevator.Ordr_Confirmed) && !Elevators[selfId].Elevator.LocalOrders[floor][btn])) {
	if shouldStartLocalOrder(assignedOrders, Elevators, activeElevators, selfId, floor, btn) {
		order := elevio.ButtonEvent{
			Floor:  floor,
			Button: elevio.ButtonType(btn),
		}
		select {
		case localAssignedOrder <- order:
			// fmt.Println("start")
			assignedOrders[selfId] = setOrder(assignedOrders[selfId], floor, btn, elevator.Ordr_Confirmed)
		default:
			//non blocking
		}

	}
	return assignedOrders
}

// func clearOrder(assignedOrders map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId, id string, floor, btn int) map[string][][]elevator.OrderState{
// 	if ordersSynced(assignedOrders, Elevators, activeElevators, selfId, id, floor, btn) && ((assignedOrders)[id][floor][btn] == elevator.Ordr_Complete) {
// 		assignedOrders[id] = setOrder(assignedOrders[id], floor, btn, elevator.Ordr_None)
// 		//kanskje sånn her. Var sånn før
// 		assignedOrders[selfId] = setOrder(assignedOrders[selfId], floor, btn, elevator.Ordr_None)
// 	}
// 	return assignedOrders
// }

// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet.
func updateFromRemoteElevator(AssignedOrders *map[string][][]elevator.OrderState, Elevators *map[string]elevator.NetworkElevator, remoteElevatorState message.ElevatorStateMsg) {
	remoteElevator := remoteElevatorState.NetworkElevator
	(*Elevators)[remoteElevatorState.Id] = remoteElevator

	_, exists := (*AssignedOrders)[remoteElevatorState.Id]
	if !exists {
		(*AssignedOrders)[remoteElevatorState.Id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
	}
}

// Checks if all active elevators in Elevators have an assignedOrders map with keys for all active elevators on nettwork
func assignedOrdersKeysCheck(Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string) bool {

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
	// for _, elev := range Elevators {
	for _, id := range activeElevators {
		elev := Elevators[id]
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
// func setOrder(assignedOrders [][]elevator.OrderState, elevId string, floor, btn int, state elevator.OrderState) [][]elevator.OrderState{
// 	copy := deepcopy.DeepCopy2DSlice(assignedOrders)
// 	copy[floor][btn] = state
// 	return copy
// }

func peerUpdateHandler(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, p peers.PeerUpdate) map[string][][]elevator.OrderState{

	//detects disconnected elevators and reassigns their orders
	if len(p.Lost) > 0 {

		//kanskje dette gjør at ordre ikke blir reassigned? Teste dette
		//tanken er at en ny heis ikke skal dø
		if assignedOrdersKeysCheck(elevators, activeElevators, selfId) {
			reassignOrdersFromDisconnectedElevators(assignedOrders, deepcopy.DeepCopyElevatorsMap(elevators),  p.Lost, activeElevators, selfId)
			// reassignOrders(deepcopy.DeepCopyElevatorsMap(*Elevators), assignedOrders, activeElevators, selfId)
			// for _, elev := range p.Lost {
			// 	for floor := range numFloors {
			// 		for btn := range numBtns - 1 {
			// 			setOrder(assignedOrders, elev, floor, btn, elevator.Ordr_Unknown)
			// 		}
			// 	}
			// }
		}
	}

	//if elevator(s) has only disconnected and reconnects it has its latest info about its order and to not override it having no order the order is set to complete
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
