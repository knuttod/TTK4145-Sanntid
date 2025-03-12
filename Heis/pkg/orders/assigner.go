package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	// "fmt"
	//"strconv"
)

func reassignOrders(elevators map[string]elevator.NetworkElevator, assignedOrders map[string][][]elevator.RequestState, reassignOrderCH chan elevio.ButtonEvent) {
	
	for _, elev := range elevators {
		if elev.Elevator.Behaviour == elevator.EB_Unavailable {
			orders := assignedOrders[elev.Elevator.Id]
			for floor := range orders{
				for button := 0; button < 2; button++ {
					if orders[floor][button] == elevator.Order || 
					orders[floor][button] == elevator.Confirmed {
						reassignOrderCH <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(button),
						}
					}
				}
			}
		}
	}
}




func assignOrder(AssignedOrders *map[string][][]elevator.RequestState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, order elevio.ButtonEvent) {
	
	
	if len((*AssignedOrders)) < 2 || order.Button == elevio.BT_Cab {
		if (*AssignedOrders)[selfId][order.Floor][order.Button] == elevator.None && ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, selfId, order.Floor, int(order.Button)){
			temp := (*AssignedOrders)[selfId]
			temp[order.Floor][order.Button] = elevator.Order
			(*AssignedOrders)[selfId] = temp
		}
		return
	}
	minCost := 99999
	 //denne må endres på, oppdaterer local orders mappet, noe den ikke skal
	elevCost := 0
	var minElev string
	for _, elev := range activeElevators{
		//temp
		if (*AssignedOrders)[elev][order.Floor][order.Button] == elevator.Confirmed {
			return
		}
		elevCost = cost(Elevators[elev].Elevator, order)
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	if (*AssignedOrders)[minElev][order.Floor][order.Button] == elevator.None && ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, minElev, order.Floor, int(order.Button)){
		temp := (*AssignedOrders)[minElev]
		temp[order.Floor][order.Button] = elevator.Order
		(*AssignedOrders)[minElev] = temp
	}
}
