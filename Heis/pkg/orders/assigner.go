package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
	"math"
	// "fmt"
	//"strconv"
)

func reassignOrders(elevators map[string]elevator.NetworkElevator, assignedOrders *map[string][][]elevator.OrderState, activeElevators []string, selfId string) {

	for id, elev := range elevators {
		if (elev.Elevator.Behaviour == elevator.EB_Unavailable) || (elev.Elevator.Obstructed) || (elev.Elevator.MotorStop) {
			orders := (*assignedOrders)[id]
			for floor := range N_floors {
				for btn := 0; btn < 2; btn++ {
					if orders[floor][btn] == elevator.Ordr_Unconfirmed ||
						orders[floor][btn] == elevator.Ordr_Confirmed {
						order := elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
						}
						fmt.Println("reassign")
						assignOrder(assignedOrders, elevators, activeElevators, selfId, order)
						//unkown ac nettet
						//complete ved obstruction eller motorstopp
						setOrder(assignedOrders, elev.Elevator.Id, floor, btn, elevator.Ordr_Complete)
					}
				}
			}
		}
	}
}

func reassignOrdersFromUnavailable(elevators map[string]elevator.NetworkElevator, assignedOrders *map[string][][]elevator.OrderState, activeElevators []string, selfId string) {

	for _, elevId := range activeElevators {
		elev := elevators[elevId].Elevator
		if elev.MotorStop || elev.Obstructed {
			orders := elevators[elevId].AssignedOrders[elevId]
			for floor := range N_floors {
				for btn := range (N_buttons - 1) {
					if orders[floor][btn] == elevator.Ordr_Unconfirmed ||
					orders[floor][btn] == elevator.Ordr_Confirmed {
						order := elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
						}
						fmt.Println("reassign from unavailable")
						assignOrder(assignedOrders, elevators, activeElevators, selfId, order)
						setOrder(assignedOrders, elevId, floor, btn, elevator.Ordr_Complete)
					}
				}
			}
		}
	}

}

func reassignOrdersFromDisconnectedElevators(elevators map[string]elevator.NetworkElevator, assignedOrders *map[string][][]elevator.OrderState, lostElevators, activeElevators []string, selfId string) {

	for _, elevId := range lostElevators {
		orders := elevators[elevId].AssignedOrders[elevId]
		for floor := range N_floors {
			for btn := range (N_buttons - 1) {
				if orders[floor][btn] == elevator.Ordr_Unconfirmed ||
					orders[floor][btn] == elevator.Ordr_Confirmed {
						order := elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
						}
						fmt.Println("reassign from disconnect")
						assignOrder(assignedOrders, elevators, activeElevators, selfId, order)
						setOrder(assignedOrders, elevId, floor, btn, elevator.Ordr_Unknown)
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

	if (len(activeElevators) < 2) || (order.Button == elevio.BT_Cab) {
		if (((*AssignedOrders)[selfId][order.Floor][order.Button] == elevator.Ordr_None) || ((*AssignedOrders)[selfId][order.Floor][order.Button] == elevator.Ordr_Unknown) || (((*AssignedOrders)[selfId][order.Floor][order.Button] == elevator.Ordr_Confirmed) && Elevators[selfId].Elevator.LocalOrders[order.Floor][order.Button])) && ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, selfId, order.Floor, int(order.Button)) {
			setOrder(AssignedOrders, selfId, order.Floor, int(order.Button), elevator.Ordr_Unconfirmed)
		}
		return
	}

	//High cost to ensure that the first elevator that is not obstructed or motorstop is chosen
	minCost := 99999
	
	elevCost := 0

	var minElev string
	for _, elev := range activeElevators {

		if (Elevators[elev].Elevator.Obstructed) || (Elevators[elev].Elevator.MotorStop) {
			continue
		}
		// //if an order is active for another elevator, do not assign it
		// if (*AssignedOrders)[elev][order.Floor][order.Button] == elevator.Ordr_Confirmed {
		// 	return
		// }
		elevCost = cost(Elevators[elev].Elevator, order)
		//Adding distance to cost for differentate between elevators with same cost
		distance := math.Abs(float64(Elevators[elev].Elevator.Floor) - float64(order.Floor))
		elevCost +=int(distance) *3

		
		
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	if (((*AssignedOrders)[minElev][order.Floor][order.Button] == elevator.Ordr_None) || ((*AssignedOrders)[minElev][order.Floor][order.Button] == elevator.Ordr_Unknown)) && ordersSynced(*AssignedOrders, Elevators, activeElevators, selfId, minElev, order.Floor, int(order.Button)) {
		setOrder(AssignedOrders, minElev, order.Floor, int(order.Button), elevator.Ordr_Unconfirmed)
	}
}
