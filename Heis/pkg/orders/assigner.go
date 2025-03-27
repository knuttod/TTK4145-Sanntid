package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"fmt"
	"math"
	
)

//reassigns orders from elevators on nettwork, with either motorstop or obstruction
func reassignOrdersFromUnavailable(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string) map[string][][]elevator.OrderState {
	for _, elevId := range activeElevators {
		elev := elevators[elevId].Elevator
		if elev.MotorStop || elev.Obstructed {
			orders := elevators[elevId].AssignedOrders[elevId]
			for floor := range numFloors {
				for btn := range numBtns - 1 {
					if orders[floor][btn] == elevator.Ordr_Unconfirmed ||
						orders[floor][btn] == elevator.Ordr_Confirmed {
						order := elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
						}
						assignedOrders = assignOrder(assignedOrders, elevators, activeElevators, selfId, order)
						//set order to complete as order is taken over by another elevator
						assignedOrders[elevId] = setOrder(assignedOrders[elevId], order.Floor, int(order.Button), elevator.Ordr_Complete)
					}
				}
			}
		}
	}
	return assignedOrders
}

//reassigns orders from elevators disconnecting from network
func reassignOrdersFromDisconnectedElevators(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, lostElevators, activeElevators []string, selfId string) map[string][][]elevator.OrderState {

	for _, elevId := range lostElevators {
		orders := elevators[elevId].AssignedOrders[elevId]
		for floor := range numFloors {
			for btn := range numBtns - 1 {
				if orders[floor][btn] == elevator.Ordr_Unconfirmed ||
					orders[floor][btn] == elevator.Ordr_Confirmed {
					order := elevio.ButtonEvent{
						Floor:  floor,
						Button: elevio.ButtonType(btn),
					}
					fmt.Println("reassign from disconnect")
					assignedOrders = assignOrder(assignedOrders, elevators, activeElevators, selfId, order)
				}
				//sets all hall orders to unkown to ensure correct merging of orders when elevator reconnects
				assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Unknown)
			}
		}
	}
	return assignedOrders
}

// Assigns an order to the elevator on the network with lowest cost. 
// Unavailable elevators (obstructed/motorstop) is not assigned any orders. 
// If another available elevator already has the order active the order is not assigned again. 
func assignOrder(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, order elevio.ButtonEvent) map[string][][]elevator.OrderState {

	if (len(activeElevators) < 2) || (order.Button == elevio.BT_Cab) {
		
		//should only take order if elevators are synced and the orders is not already started/taken
		if ordersSynced(assignedOrders, elevators, activeElevators, selfId, order.Floor, int(order.Button)) && 
		((assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_None) || 
		(assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_Unknown)){
			assignedOrders[selfId] = setOrder(assignedOrders[selfId], order.Floor, int(order.Button), elevator.Ordr_Unconfirmed)
		}
		return assignedOrders
	}

	
	minCost := 99999
	elevCost := 0
	// sets self as min to prevent index error if no elevators are available
	minElev := selfId

	for _, elev := range activeElevators {
		
		//unaivalable elevators should not be assigned orders
		if (elevators[elev].Elevator.Obstructed) || (elevators[elev].Elevator.MotorStop) {
			continue
		}

		// Checks if order is already taken by another elevator. In this case the order should not be assigned again
		// This check is after the unavailable elevators to prevent not taking orders from these elevators. 
		activeOrders := elevators[elev].AssignedOrders[elev]
		if (activeOrders[order.Floor][order.Button] == elevator.Ordr_Unconfirmed) || (activeOrders[order.Floor][order.Button] == elevator.Ordr_Confirmed) {
			return assignedOrders
		}

		elevCost = cost(elevators[elev].Elevator)

		//Adding distance to cost to differentate between elevators with same cost
		distance := math.Abs(float64(elevators[elev].Elevator.Floor - order.Floor))
		elevCost += int(distance) *3


		//choose elevator with lower cost
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}

	//should only take order if elevators are synced and the orders is not already started/taken
	if ordersSynced(assignedOrders, elevators, activeElevators, minElev, order.Floor, int(order.Button)) && 
		((assignedOrders[minElev][order.Floor][order.Button] == elevator.Ordr_None) || 
		(assignedOrders[minElev][order.Floor][order.Button] == elevator.Ordr_Unknown)){
		assignedOrders[minElev] = setOrder(assignedOrders[minElev], order.Floor, int(order.Button), elevator.Ordr_Unconfirmed)
	}
	return assignedOrders
}

//ensure that the elevator struct given as input is a deepcopy as this function changes the values

//Calculates the cost of an elevator for the given order
func cost(elev elevator.Elevator) int {

		duration := 0

		switch elev.Behaviour {
		//If elevator is idle, and there is no given direction, there is no extra cost
		case elevator.EB_Idle:
			directionAndBehaviour := fsm.ChooseDirection(elev)
			elev.Dirn = directionAndBehaviour.Dirn
			elev.Behaviour = directionAndBehaviour.Behaviour
			if elev.Dirn == elevio.MD_Stop {
				return duration 
				
			}
		case elevator.EB_Moving:
			//If elevator is moving, we add the time it takes to reach the floor
			duration += travelTime / 2 
			elev.Floor += int(elev.Dirn)
		case elevator.EB_DoorOpen:
			//Subtracting the time it takes to open the door, since the elevator is already idle with the door open
			duration -= int(fsm.DoorTimerInterval.Seconds())
		}
		for duration < 999{

			// An elevator should not be moving 
			if elev.Floor < 0 || elev.Floor > (numFloors - 1) {
				break
			}

			//Returning the duration when the elevator should stop
			if fsm.ShouldStop(elev) {
				elev = costClearAtCurrentFloor(elev)
				duration += int(fsm.DoorTimerInterval.Seconds())
				
				directionAndBehaviour := fsm.ChooseDirection(elev)
				elev.Dirn = directionAndBehaviour.Dirn
				elev.Behaviour = directionAndBehaviour.Behaviour
				if elev.Dirn == elevio.MD_Stop {
					return duration 
				}
			}
			//Adding the time it takes to reach the next floor, considering its direction and travel time
			elev.Floor += int(elev.Dirn)
			duration += travelTime       
		}
	//Return high cost if elevator is unavailable
	return 999 
}

// Same version as in fsm, but without sending on completed channel to orders
func costClearAtCurrentFloor(elev elevator.Elevator) elevator.Elevator {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		for btn := 0; btn < numBtns; btn++ {
			elev.LocalOrders[elev.Floor][btn] = false
		}
	case elevator.CV_InDirn:
		elev.LocalOrders[elev.Floor][elevio.BT_Cab] = false
		switch elev.Dirn {
		case elevio.MD_Up:
			if (!fsm.LocalOrderAbove(elev)) && !(elev.LocalOrders[elev.Floor][elevio.BT_HallUp]) {
				elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
			}
			elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if (!fsm.LocalOrderBelow(elev)) && !(elev.LocalOrders[elev.Floor][elevio.BT_HallDown]) {
				elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
			}
			elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
		case elevio.MD_Stop:
			elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
			elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
		}
	}
	return elev
}