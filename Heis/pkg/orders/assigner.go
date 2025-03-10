package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
	// "fmt"
	"strconv"
)

func ReassignOrders(elevators []*elevator.Elevator, localRequest chan elevio.ButtonEvent) {
	lowestID := 9999

	//Finn den tilgjengelige heisen med lavest ID
	for _, elev := range elevators {
		if elev.Behaviour != elevator.EB_Unavailable {
			ID, _ := strconv.Atoi(elev.Id)
			if ID < lowestID {
				lowestID = ID
			}
		}
	}

	//Itererer gjennom alle heisene for å reassigne ordrer fra unavailable heiser
	for _, elev := range elevators {
		if elev.Behaviour == elevator.EB_Unavailable{
			//Henter matrisen fra assignedORders med heisens egen ID som key
			orders, ok := elev.AssignedOrders[elev.Id]
			if !ok{
				continue //ingen bestillinger for denne heisen
			}

			for floor := range orders{
				for button := 0; button < len(orders[floor])-1; button++ {
					
					//Sjekk om requestStaten er enten Order eller Comfirmed (Trenger kanskje bare comfirmed)
					if orders[floor][button] == elevator.Order ||
						orders[floor][button] == elevator.Confirmed {

						//Hvis den lokale heisen har lavest ID, ta over orderen 
						if elevators[0].Id == strconv.Itoa(lowestID) {
							localRequest <- elevio.ButtonEvent{
								Floor:  floor,
								Button: elevio.ButtonType(button),
							}
						}
					}
				}
			}
		}
	}
}




func assignOrder(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, order elevio.ButtonEvent) {
	
	
	if len((*e).AssignedOrders) < 2 || order.Button == elevio.BT_Cab {
		if (*e).AssignedOrders[(*e).Id][order.Floor][order.Button] == elevator.None && ordersSynced(*e, remoteElevators, activeElevators, (*e).Id, order.Floor, int(order.Button)){
			temp := (*e).AssignedOrders[(*e).Id]
			temp[order.Floor][order.Button] = elevator.Order
			(*e).AssignedOrders[(*e).Id] = temp
		}
		return
	}
	minCost := 99999
	elev := *e
	elevCost := cost(deepCopyElev(elev), order)  //denne må endres på, oppdaterer local orders mappet, noe den ikke skal
	minCost = elevCost
	// elevCost := 0
	minElev := (*e).Id
	// for id, elev := range remoteElevators {
	// 	elevCost = cost(elev, order)
	// 	if elevCost < minCost {
	// 		minCost = elevCost
	// 		minElev = id
	// 	}
	// }
	for _, elev := range activeElevators{

		if elev == (*e).Id {
			continue
		}

		elevCost = cost(remoteElevators[elev], order)
		
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
		fmt.Println("Duration: ", minElev, ", ", minCost)

	}
	
	if (*e).AssignedOrders[minElev][order.Floor][order.Button] == elevator.None && ordersSynced(*e, remoteElevators, activeElevators, minElev, order.Floor, int(order.Button)){
		temp := (*e).AssignedOrders[minElev]
		temp[order.Floor][order.Button] = elevator.Order
		(*e).AssignedOrders[minElev] = temp

	}
	
}
func deepCopyElev(e elevator.Elevator) elevator.Elevator {
    newElev := e
    newElev.AssignedOrders = make(map[string][][]elevator.RequestState)
    for key, orders := range e.AssignedOrders {
        // orders er en 2D-slice: [][]elevator.RequestState
        newOrders := make([][]elevator.RequestState, len(orders))
        for i, row := range orders {
            newRow := make([]elevator.RequestState, len(row))
            copy(newRow, row)
            newOrders[i] = newRow
        }
        newElev.AssignedOrders[key] = newOrders
    }
    return newElev
}
