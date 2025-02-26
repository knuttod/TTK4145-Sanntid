package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgtypes"
	"fmt"
	"Heis/pkg/fsm"
)


func RequestMerger(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, stateUpdated chan bool, startOrder chan chan elevio.ButtonEvent) {
	var currentState elevator.RequestState
	var updateState elevator.RequestState

	for {
		select {
		case <- stateUpdated:
			for id, _ := range (*e).AssginedOrders {
				for _, elev := range (*e).AssginedOrders {
					for floor := range N_floors{
					// for Requests
					// for btn := range (*e).Requests[floor] {
					// 	currentState = (*e).Requests[floor][btn]
					// 	updateState = elev.Requests[floor][btn]
					// 	if updateState - currentState == 1 || (updateState == 0 && currentState == 3) {
					// 		(*e).Requests[floor][btn] = updateState
					// 	}
					// 	confirmOrCloseRequests(e, remoteElevators, floor, btn)
					// }

					// for AssginedOrders
					for btn := range N_buttons {
							currentState = (*e).AssginedOrders[id][floor][btn]
							updateState = elev.AssginedOrders[id][floor][btn]
							if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
								(*e).AssginedOrders[(*e).Id][floor][btn] = updateState
							}
						}	
					}
					confirmOrCloseOrders(e, remoteElevators, id, floor, btn)
				}
			}

			for floor := range N_floors {
				for btn := range N_buttons {
					shouldStartOrder(e, startOrder, (*e).id, floor, btn)
				}
			}
		}
	}
}

// func requestSynced(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, floor, btn int) bool {
// 	for _, elev := range remoteElevators {
// 		if (e*).requests[floor][btn] != elev.requests[floor][btn] {
// 			return false
// 		}
// 	}
// 	return true
// }
// func confirmOrCloseRequests(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, floor, btn int) {
// 	if requestSynced(e, remoteElevators, floor, btn) {
// 		if updateState == 1 {
// 			(*e).Requests[floor][btn] = 2
// 		}
// 		if updateState == 3 {
// 			(*e).Requests[floor][btn] = 0
// 		}
// 	}
// }

func ordersSynced(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, id, floor, btn int) bool {
	for _, elev := range remoteElevators {
		if (e*).AssignedOrders[id][floor][btn] != elev.AssignedOrders[id][floor][btn] {
			return false
		}
	}
	return true
}

func confirmOrCloseOrders(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, id, floor, btn int) {
	if ordersSynced(e, remoteElevators, id, floor, btn) {
		if updateState == 1 {
			(e*).AssignedOrders[id][floor][btn] = 2
		}
		if updateState == 3 {
			(e*).AssignedOrders[id][floor][btn] = 0
		}
	}
}


func shouldStartOrder(e* elevator.Elevator, startOrder chan elevio.ButtonEvent, id, floor, btn int) {
	// for floor := range N_floors {
	// 	for btn := range N_buttons {
	// 		if (*e).AssginedOrders[(*e).Id][floor][btn] {
	// 			startOrder <- elevio.ButtonEvent{
	// 				Floor:  floor,
	// 				Button: elevio.ButtonType(button)
	// 			}
	// 		}
	// 	}
	// }
	if ordersSynced(e, remoteElevators, id, button_input.floor, button_input.btn) && (*e).AssginedOrders[id][floor][btn] == elevator.Confirmed{
		startOrder <- elevio.ButtonEvent{
			Floor:  floor,
			Button: elevio.ButtonType(button)
		}
	}
}

// func OrderDistributer(e *elevator.Elevator, LocalOrderOut chan elevio.ButtonEvent) {
// 	//egentlig bruke kostfunksjoner her

// 	//Basicly sjekke om alle ordre er synkronisert, om det er tilfelle kan man fordele uncofirmed orders.
// 	// Dette gjør man ved å velge den heisen som har lavest kost for ordren fra kostfunksjonene. 
// 	// Heisen som får lavest kost selv setter ordren til å bli gjort lokalt og endrer fra uncofirmed til confirmed i ordrematrisen
// 	var state int
// 	if (*e).Index == 1 {
// 		for floor := range (*e).LocalOrders {
// 			for btn := range (*e).LocalOrders[floor] {
// 				if btn == 2 { //Cab call
// 					state = (*e).GlobalOrders[floor][(*e).Index + 1]
// 					if GlobalOrderSynced(e, state, floor, btn) && state == 1 {
// 						(*e).GlobalOrders[floor][(*e).Index + 1] = 2
// 						// send button input to FSM
// 						LocalOrderOut <- elevio.ButtonEvent{Floor : floor, Button : elevio.ButtonType(btn)}
// 					}
// 				} else {
// 					state = (*e).GlobalOrders[floor][btn]
// 					if GlobalOrderSynced(e, state, floor, btn) && state == 1 {
// 						(*e).GlobalOrders[floor][btn] = 2
// 						// send button input to FSM
// 						LocalOrderOut <- elevio.ButtonEvent{Floor : floor, Button : elevio.ButtonType(btn)}
// 					}
// 				}
				
// 			}
// 		}
// 	}
// }

// func LocalButtonPressHandler (e *elevator.Elevator, drv_buttons chan elevio.ButtonEvent, localRequest chan elevator.Order) {
// 	for {
		
// 		button_input := <- drv_buttons
// 		Order := elevator.Order { 
// 			State : 1,
// 			Action: button_input,
// 		}
		
// 		// Do not need this, but does not create any problems i think? May be bad code quality 
// 		// May be wrong if orders is to be assigned to another elevator
// 		(*e).LocalOrders[button_input.Floor][button_input.Button] = 1

// 		localRequest <- Order
// 	}
// }
