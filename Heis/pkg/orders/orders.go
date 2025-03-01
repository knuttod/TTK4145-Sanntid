package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
	// "fmt"
	// "Heis/pkg/msgTypes"
	// "fmt"
)

const N_floors = 4
const N_buttons = 3

func orderMerger(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator) {
	var currentState elevator.RequestState
	var updateState elevator.RequestState

	// for id, _ := range remoteElevators{
	for id, _ := range (*e).AssignedOrders {	
		for floor := range N_floors{
			for btn := range N_buttons {
				for _, elev := range remoteElevators {
					// fmt.Println(id)
					currentState = (*e).AssignedOrders[id][floor][btn]
					updateState = elev.AssignedOrders[id][floor][btn]

					if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
						temp := (*e).AssignedOrders[id]
						temp[floor][btn] = updateState
						(*e).AssignedOrders[id] = temp
						// (*e).AssignedOrders[(*e).Id][floor][btn] = updateState
					}
				}	
				confirmOrCloseOrders(e, remoteElevators, id, floor, btn)
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

func ordersSynced(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, id string, floor, btn int) bool {
	for _, elev := range remoteElevators {
		if (*e).AssignedOrders[id][floor][btn] != elev.AssignedOrders[id][floor][btn] {
			return false
		}
	}
	return true
}

func confirmOrCloseOrders(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, id string, floor, btn int) {
	if ordersSynced(e, remoteElevators, id, floor, btn) {
		if  (*e).AssignedOrders[id][floor][btn] == elevator.Order {
			temp := (*e).AssignedOrders[id]
			temp[floor][btn] = elevator.Confirmed
			(*e).AssignedOrders[id] = temp
			fmt.Println("Confirmed")
			// (*e).AssignedOrders[id][floor][btn] = elevator.Confirmed
		}
		if  (*e).AssignedOrders[id][floor][btn] == elevator.Complete{
			temp := (*e).AssignedOrders[id]
			temp[floor][btn] = elevator.None
			(*e).AssignedOrders[id] = temp
			// (*e).AssignedOrders[id][floor][btn] = elevator.None
		}
	}
}


func shouldStartLocalOrder(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, id string, floor, btn int) bool {
	return ordersSynced(e, remoteElevators, id, floor, btn) && (*e).AssignedOrders[id][floor][btn] == elevator.Confirmed
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




func OrderHandler(e *elevator.Elevator, remoteElevators *map[string]elevator.Elevator, localAssignedOrder, localRequest chan elevio.ButtonEvent, externalStateUpdate chan bool) {
	var activeLocalOrders [N_floors][N_buttons]bool
	for {
		select {
		case btn_input := <- localRequest:
			assignOrder(e, *remoteElevators, btn_input)
			// temp := (*e).AssignedOrders["heis1"]
			// temp[btn_input.Floor][btn_input.Button] = elevator.Order
			// (*e).AssignedOrders["heis1"] = temp
		case <- externalStateUpdate:
			//funksjon for å opdatere remoteElevators og assignedElevators
			orderMerger(e, *remoteElevators)
		}
		// case for disconnection or timout for elevator to reassign orders

		// Check if an unstarted assigned order should be started
		for floor := range N_floors {
			for btn := range N_buttons {
				if (*e).AssignedOrders[(*e).Id][floor][btn] != elevator.Confirmed {
					activeLocalOrders[floor][btn] = false
				}
				if shouldStartLocalOrder(e, *remoteElevators, (*e).Id, floor, btn) && !activeLocalOrders[floor][btn] {
					localAssignedOrder <- elevio.ButtonEvent{
						Floor:  floor,
						Button: elevio.ButtonType(btn),
					}
					activeLocalOrders[floor][btn] = true
					// Må bare passe på at en ordre ikke blir sent hele tiden
					fmt.Println("Order")
				}
			}
		}
	}
}