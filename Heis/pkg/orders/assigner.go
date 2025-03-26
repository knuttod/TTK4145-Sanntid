package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"fmt"
	"math"
	// "fmt"
	//"strconv"
)
//reassigns orders from elevators on nettwork with either motorstop or obstruction
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
				//sets all hall orders to unkwon to ensure correct merging of orders when elevator reconnects
				assignedOrders[elevId] = setOrder(assignedOrders[elevId], floor, btn, elevator.Ordr_Unknown)
			}
		}
	}
	return assignedOrders
}

func assignOrder(assignedOrders map[string][][]elevator.OrderState, elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, order elevio.ButtonEvent) map[string][][]elevator.OrderState {

	if (len(activeElevators) < 2) || (order.Button == elevio.BT_Cab) {
		
		//should only take order if elevators are synced and the orders is not already taken
		if ordersSynced(assignedOrders, elevators, activeElevators, selfId, order.Floor, int(order.Button)) && 
		((assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_None) || 
		(assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_Unknown)){
		// !((assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_Complete) ||
		// ((assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_Confirmed) && elevators[selfId].Elevator.LocalOrders[order.Floor][order.Button])){
		// ((assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_None) || 
		// (assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_Unknown) || 
		// ((assignedOrders[selfId][order.Floor][order.Button] == elevator.Ordr_Confirmed) && elevators[selfId].Elevator.LocalOrders[order.Floor][order.Button])) {
			assignedOrders[selfId] = setOrder(assignedOrders[selfId], order.Floor, int(order.Button), elevator.Ordr_Unconfirmed)
		}
		return assignedOrders
	}

	//High cost to ensure that an elevator that is not obstructed or motorstop is chosen
	minCost := 99999
	elevCost := 0

	// sets self as min to prevent index error if no elevators are available
	minElev := selfId

	for _, elev := range activeElevators {

		//unaivalable elevators should not be assigned orders
		if (elevators[elev].Elevator.Obstructed) || (elevators[elev].Elevator.MotorStop) {
			continue
		}
		elevCost = cost(elevators[elev].Elevator)
		//Adding distance to cost to differentate between elevators with same cost
		distance := math.Abs(float64(elevators[elev].Elevator.Floor) - float64(order.Floor))
		elevCost += int(distance) * 3

		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	if ordersSynced(assignedOrders, elevators, activeElevators, minElev, order.Floor, int(order.Button)) && 
		((assignedOrders[minElev][order.Floor][order.Button] == elevator.Ordr_None) || 
		(assignedOrders[minElev][order.Floor][order.Button] == elevator.Ordr_Unknown)){
	// if ((assignedOrders[minElev][order.Floor][order.Button] == elevator.Ordr_None) || (assignedOrders[minElev][order.Floor][order.Button] == elevator.Ordr_Unknown)) && ordersSynced(assignedOrders, elevators, activeElevators, selfId, minElev, order.Floor, int(order.Button)) {
		assignedOrders[minElev] = setOrder(assignedOrders[minElev], order.Floor, int(order.Button), elevator.Ordr_Unconfirmed)
	}
	return assignedOrders
}


//ensure that the elevator struct given as input is a deepcopy as this function changes the values
func cost(elev elevator.Elevator) int {
	if elevator.ElevatorBehaviour(elev.Behaviour) != elevator.EB_Unavailable {

		duration := 0


		switch elev.Behaviour {
		case elevator.EB_Idle:
			directionAndBehaviour := fsm.ChooseDirection(elev)
			elev.Dirn = directionAndBehaviour.Dirn
			elev.Behaviour = directionAndBehaviour.Behaviour
			if elev.Dirn == elevio.MD_Stop {
				return duration //Dersom EB_IDLE, og hvis det er ingen retning, blir det ingen ekstra kostnad

			}
		case elevator.EB_Moving:
			duration += travelTime / 2 //dersom heisen er i beveglse legger vi til en kostand
			elev.Floor += int(elev.Dirn)
		case elevator.EB_DoorOpen:
			//Trekker fra kostnad siden heisen allerede står i ro med dørene åpne og er dermed:
			//Klar til å ta imot nye bestillinger på denne etasjoen, uten ekstra (halvparten) ventetid for å åpne dører
			duration -= int(elev.Config.DoorOpenDuration_s)

		}
		for duration < 999 {
			// An elevator should not be moving 
			if elev.Floor < 0 || elev.Floor > (numFloors - 1) {
				break
			}
			if fsm.ShouldStop(elev) {
				elev = costClearAtCurrentFloor(elev)
				duration += int(fsm.DoorTimerInterval)
				// might not need this?
				directionAndBehaviour := fsm.ChooseDirection(elev)
				elev.Dirn = directionAndBehaviour.Dirn
				elev.Behaviour = directionAndBehaviour.Behaviour
				// ...
				if elev.Dirn == elevio.MD_Stop {
					return duration //returner duration når den simulerte heisen har kommet til en stopp
				}
			}
			elev.Floor += int(elev.Dirn) //Hvis det ikke er kommet noe tegn på at den stopper sier vi at den estimerte heisen sier vi her at den går til en ny etasje
			duration += travelTime       //da vil vi også legge til en TRAVEL_TIME kostand for denne opeerasjonen
		}
		//		return 999

	}
	return 999 //returnerer høy kostnad dersom heisen er EB_unavailable
}

// version without sending on completed channel to orders
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
			if (!fsm.LocalOrderAbove(elev)) && (elev.LocalOrders[elev.Floor][elevio.BT_HallUp] == false) {
				elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
			}
			elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if (!fsm.LocalOrderBelow(elev)) && (elev.LocalOrders[elev.Floor][elevio.BT_HallDown] == false) {
				elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
			}
			elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
		// case elevio.MD_Stop:
		// 	elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
		// 	elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
		// 	elev.LocalOrders[elev.Floor][elevio.BT_Cab] = false
		default:
			elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
			elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
			//elev.LocalOrders[elev.Floor][elevio.BT_Cab] = false
		}
	default:

	}
	return elev
}