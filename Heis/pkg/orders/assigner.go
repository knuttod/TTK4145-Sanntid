package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	// "fmt"
	//"strconv"
)

// func ReassignOrders(elevators []*elevator.Elevator, ch_newLoacalOrder chan elevio.ButtonEvent) {
// 	lowestID := 999

// 	for _, elev := range elevators {
// 		if elev.Behave != elevator.Unavailable {
// 			ID, _ := strconv.Atoi(elev.ID)
// 			if ID < lowestID {
// 				lowestID = ID
// 			}
// 		}
// 	}
// 	for _, elev := range elevators {
// 		if elev.Behave == elevator.Unavailable {
// 			for floor := range elev.(*e).AssignedOrders {
// 				for button := 0; button < 2; button++ {
// 					// if elev.AssignedOrders[floor][button] == elevator.Order ||
// 					if elev.AssignedOrders[elev.Id][floor][button] == elevator.Confirmed {
// 						if elevators[elevator.LocalElevator].ID == strconv.Itoa(lowestID) {
// 							ch_newLoacalOrder <- elevio.ButtonEvent{
// 								Floor:  floor,
// 								Button: elevio.ButtonType(button),
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }




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
