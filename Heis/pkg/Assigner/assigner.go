package assigner

import (
	"Heis/pkg/Assigner/cost"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"

	//"fmt"
	"strconv"
)

func ReassignOrders(elevators []*elevator.DistributorElevator, ch_newLoacalOrder chan elevio.ButtonEvent) {
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
			for floor := range elev.Requests {
				for button := 0; button < len(elev.Requests[floor])-1; button++ {
					if elev.Requests[floor][button] == elevator.Order ||
						elev.Requests[floor][button] == elevator.Comfirmed {
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

func AssignOrder(elevators []*elevator.DistributorElevator, order elevio.ButtonEvent) {
	if len(elevators) < 2 || order.Button == elevio.BT_Cab {
		elevators[elevator.LocalElevator].Requests[order.Floor][order.Button] = elevator.Order
		return
	}
	minCost := 99999
	elevCost := 0
	var minElev *elevator.DistributorElevator
	for _, elev := range elevators {
		elevCost = cost.Cost(elev, order)
		if elevCost < minCost {
			minCost = elevCost
			minElev = elev
		}
	}
	(*minElev).Requests[order.Floor][order.Button] = elevator.Order
}
