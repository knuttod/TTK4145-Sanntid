package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	// "fmt"
	//"strconv"
)

func reassignOrders(elevators map[string]elevator.NetworkElevator, assignedOrders map[string][][]elevator.OrderState, reassignOrderCH chan elevio.ButtonEvent) {
	
	for _, elev := range elevators {
		if elev.Elevator.Behaviour == elevator.EB_Unavailable {
			orders := assignedOrders[elev.Elevator.Id]
			for floor := range orders{
				for button := 0; button < 2; button++ {
					if orders[floor][button] == elevator.Ordr_Unconfirmed || 
					orders[floor][button] == elevator.Ordr_Confirmed {
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


// //For å passe på at man ikke endrer på slicen. Slices er tydelighvis by reference, forstår ikke helt, men er det som er feilen
// copy := make([][]bool, len(Elevators[selfId].Elevator.LocalOrders))
// for i := range copy {
// 	copy[i] = append([]bool(nil), Elevators[selfId].Elevator.LocalOrders[i]...) // Ensure deep copy
// }
// temp := Elevators[selfId]
// temp.Elevator.LocalOrders = copy
// Elevators[selfId] = temp

func assignOrder(AssignedOrders *map[string][][]elevator.OrderState, Elevators map[string]elevator.NetworkElevator, activeElevators []string, selfId string, order elevio.ButtonEvent) {
	
	
	if len((*AssignedOrders)) < 2 || order.Button == elevio.BT_Cab {
		if (((*AssignedOrders)[selfId][order.Floor][order.Button] == elevator.Ordr_None) || ((*AssignedOrders)[selfId][order.Floor][order.Button] == elevator.Ordr_Unknown)) && ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, selfId, order.Floor, int(order.Button)){
			temp := (*AssignedOrders)[selfId]
			temp[order.Floor][order.Button] = elevator.Ordr_Unconfirmed
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
		if (*AssignedOrders)[elev][order.Floor][order.Button] == elevator.Ordr_Confirmed {
			return
		}
		elevCost = cost(Elevators[elev].Elevator, order)
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	if (((*AssignedOrders)[minElev][order.Floor][order.Button] == elevator.Ordr_None) || ((*AssignedOrders)[minElev][order.Floor][order.Button] == elevator.Ordr_Unknown)) && ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, minElev, order.Floor, int(order.Button)){
		temp := (*AssignedOrders)[minElev]
		temp[order.Floor][order.Button] = elevator.Ordr_Unconfirmed
		(*AssignedOrders)[minElev] = temp
	}
}
