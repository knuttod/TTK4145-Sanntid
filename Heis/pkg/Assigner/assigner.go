package assigner

import (
	"Heis/pkg/Assigner/cost"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"

	//"fmt"
	"strconv"
)

func ReassignOrders(elevators []*elevator.Elevator, ch_newLoacalOrder chan elevio.ButtonEvent) {
	lowestID := 999

	for _, elev := range elevators {
		if elev.Behave != elevator.Unavailable {
			ID, _ := strconv.Atoi(elev.ID)
			if ID < lowestID {
				lowestID = ID
			}
		}
	}
	for _, elev := range elevators {
		if elev.Behave == elevator.Unavailable {
			for floor := range elev.AssignedOrders {
				for button := 0; button < len(elev.AssignedOrders[floor])-1; button++ {
					if elev.AssignedOrders[floor][button] == elevator.Order ||
						elev.AssignedOrders[floor][button] == elevator.Comfirmed {
						if elevators[elevator.LocalElevator].ID == strconv.Itoa(lowestID) {
							ch_newLoacalOrder <- elevio.ButtonEvent{
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

func AssignOrder(elevators []*elevator.Elevator, order elevio.ButtonEvent) *elevator.Elevator{
	if len(elevators) < 2 || order.Button == elevio.BT_Cab {
		elevators[elevator.LocalElevator].AssignedOrders[order.Floor][order.Button] = elevator.Order
		return
	}
	minCost := 99999
	elevCost := 0
	var minElev *elevator.Elevator
	for _, elev := range elevators {
		elevCost = cost.Cost(elev, order)
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	(*minElev).AssignedOrders[order.Floor][order.Button] = elevator.Order
	//(*minElev).Requests[order.Floor][order.Button] = elevator.Order
	return minElev
}
