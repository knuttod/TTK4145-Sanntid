package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"fmt"
	"Heis/pkg/fsm"
)

func GlobalOrderMerger(e *elevator.Elevator, stateRX, stateTx chan msgTypes.ElevatorStateMsg, localOrderIn chan elevator.Order, localOrderOut chan elevio.ButtonEvent) {
	var external_elevator elevator.Elevator
	var currentState int
	var updateState int

	// dummy variable for now
	var equal bool
	for {
		equal = true
		select {
		case a := <-stateRX:

			// SyncGlobalWithLocalOrder(e)
			
			external_elevator = a.Elevator
			if a.Id != (*e).Id { //ignores messages from itself and if orders are equal
				for floor := range (*e).GlobalOrders {
					for btn := range (*e).GlobalOrders[floor] {
						// If cyclic counter is 1 behind external it can update
						currentState = (*e).GlobalOrders[floor][btn]
						updateState = external_elevator.GlobalOrders[floor][btn]
						
						if updateState - currentState == 1 || (updateState == 0 && currentState == 2) {
							(*e).GlobalOrders[floor][btn] = updateState
							equal = false
						}
					}
				}
			
				// if true {//!equal {
				// 	// Må kanskje fikse funksjon for å få denne enkel eller endre på funksjonen
				// 	//OrderDistributer(e, localOrderOut)
				// 	fsm.SetAllLightsOrder((*e).GlobalOrders, e)
				// }
			}
		case a := <-localOrderIn:

			//SyncGlobalWithLocalOrder(e)

			// Kanskje sjekke om alle er på samme state før man kan gjøre dette?
			// Kan også hende at lokale of globale ordre ikke er syncet enda og knappen ikke vil lyse ved trykk/orderen ikke blir opdatert
			updateState = a.State
			if a.Action.Button == 2 {
				currentState = (*e).GlobalOrders[a.Action.Floor][2+(*e).Index-1]
			} else {
				currentState = (*e).GlobalOrders[a.Action.Floor][a.Action.Button]
			}

			if updateState - currentState == 1 || (updateState == 0 && currentState == 2) {
				(*e).GlobalOrders[a.Action.Floor][a.Action.Button] = updateState
				equal = false
			}
			OrderDistributer(e, localOrderOut)
		}

		if !equal {
			fsm.SetAllLightsOrder((*e).GlobalOrders, e)
		}
		

	}
}

// Tror bare vi trenger denne for å synkronisere når ordre er ferdige. 
// func SyncGlobalWithLocalOrders(e *elevator.Elevator) {
// 	var currentState int
// 	var updateState int
// 	for {
// 		for floor := range (*e).LocalOrders {
// 			for btn := range (*e).LocalOrders[floor] {
// 				// If cyclic counter is 1 behind external it can update
// 				updateState = (*e).LocalOrders[floor][btn]
// 				if btn == 2 {
// 					currentState = (*e).GlobalOrders[floor][(*e).Index + 1]
// 					if updateState - currentState == 1 || (updateState == 0 && currentState == 2) {
// 						(*e).GlobalOrders[floor][(*e).Index + 1] = updateState
// 					}
// 				} else {
// 					currentState = (*e).GlobalOrders[floor][btn]
// 					if updateState - currentState == 1 || (updateState == 0 && currentState == 2) {
// 						(*e).GlobalOrders[floor][btn] = updateState
// 					}
// 				}
// 			}
// 		}
// 	}
// }

func CheckForCompletedOrders(e *elevator.Elevator, localOrderIn chan elevator.Order) {
	var currentState int
	var updateState int
	for {
		for floor := range (*e).LocalOrders {
			for btn := range (*e).LocalOrders[floor] {
				// If cyclic counter is 1 behind external it can update
				updateState = (*e).LocalOrders[floor][btn]
				if btn == 2 {
					currentState = (*e).GlobalOrders[floor][(*e).Index + 1]
					if (updateState == 0 && currentState == 2) {
						//(*e).GlobalOrders[floor][(*e).Index + 1] = updateState
						localOrderIn <- elevator.Order{State: 0, Action: elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}}

					}
				} else {
					currentState = (*e).GlobalOrders[floor][btn]
					if (updateState == 0 && currentState == 2) {
						//(*e).GlobalOrders[floor][btn] = updateState
						localOrderIn <- elevator.Order{State: 0, Action: elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}}
					}
				}
			}
		}
	}
}

func GlobalOrderSynced(e *elevator.Elevator, state, floor, btn int) bool { //egentlig ta inn informasjonen fra alle heiser
	// for i := 0; i< N_elevators; i++ {
	// 	check if all active elevators have the order synced
	// }
	//return true or false
	return true
}

func OrderDistributer(e *elevator.Elevator, LocalOrderOut chan elevio.ButtonEvent) {
	//egentlig bruke kostfunksjoner her

	//Basicly sjekke om alle ordre er synkronisert, om det er tilfelle kan man fordele uncofirmed orders.
	// Dette gjør man ved å velge den heisen som har lavest kost for ordren fra kostfunksjonene. 
	// Heisen som får lavest kost selv setter ordren til å bli gjort lokalt og endrer fra uncofirmed til confirmed i ordrematrisen
	var state int
	if (*e).Index == 1 {
		for floor := range (*e).LocalOrders {
			for btn := range (*e).LocalOrders[floor] {
				if btn == 2 { //Cab call
					state = (*e).GlobalOrders[floor][(*e).Index + 1]
					if GlobalOrderSynced(e, state, floor, btn) && state == 1 {
						(*e).GlobalOrders[floor][(*e).Index + 1] = 2
						// send button input to FSM
						fmt.Println("send")
						LocalOrderOut <- elevio.ButtonEvent{Floor : floor, Button : elevio.ButtonType(btn)}
					}
				} else {
					state = (*e).GlobalOrders[floor][btn]
					if GlobalOrderSynced(e, state, floor, btn) && state == 1 {
						(*e).GlobalOrders[floor][btn] = 2
						// send button input to FSM
						fmt.Println("send2")
						LocalOrderOut <- elevio.ButtonEvent{Floor : floor, Button : elevio.ButtonType(btn)}
					}
				}
				
			}
		}
	}
}

func LocalButtonPressHandler (e *elevator.Elevator, drv_buttons chan elevio.ButtonEvent, localRequest chan elevator.Order) {
	for {
		
		button_input := <- drv_buttons
		fmt.Println("Floor:", button_input.Floor)
		fmt.Println("Button:", button_input.Button)
		Order := elevator.Order { 
			State : 1,
			Action: button_input,
		}
		
		// Do not need this, but does not create any problems i think? May be bad code quality 
		(*e).LocalOrders[button_input.Floor][button_input.Button] = 1

		localRequest <- Order
	}
}
