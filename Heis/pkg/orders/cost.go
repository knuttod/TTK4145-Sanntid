package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
)

const TRAVEL_TIME = 10
const NumElevators = 4

func cost(elev elevator.Elevator, req elevio.ButtonEvent) int {
	if elevator.ElevatorBehaviour(elev.Behaviour) != elevator.EB_Unavailable {
		// e := new(elevator.Elevator)
		// *e = *elev //lager en kopi av heisen for å estimere kjøretiden når man legger til den nye orderen

		// Siden maps i go bare har shallow copy vil endringer av aksessering av value ikke endre den faktiske i mapet
		e := elev
		e.Orders[req.Floor][req.Button] = elevator.Confirmed

		duration := 0

		switch e.Behaviour {
		case elevator.EB_Idle:
			pair := fsm.ChooseDirection(e)
			e.Dirn = pair.Dirn
			e.Behaviour = pair.Behaviour
			//requestChooseDirnection(e)
			if e.Dirn == elevio.MD_Stop {
				return duration //Dersom EB_IDLE, og hvis det er ingen retning, blir det ingen ekstra kostnad

			}
		case elevator.EB_Moving:
			duration += TRAVEL_TIME / 2 //dersom heisen er i beveglse legger vi til en kostand
			e.Floor += int(e.Dirn)
		case elevator.EB_DoorOpen:
			duration -= int(e.Config.DoorOpenDuration_s) / 2
			//Trekker fra kostnad siden heisen allerede står i ro med dørene åpne og er dermed:
			//Klar til å ta imot nye bestillinger på denne etasjoen, uten ekstra (halvparten) ventetid for å åpne dører

		}
		for {
			if fsm.ShouldStop(e) {
				e = fsm.ClearAtCurrentFloor(e)
				duration += int(e.Config.DoorOpenDuration_s)
				pair := fsm.ChooseDirection(e)
				e.Dirn = pair.Dirn
				e.Behaviour = pair.Behaviour
				if e.Dirn == elevio.MD_Stop {
					return duration //returner duration når den simulerte heisen har kommet til en stopp
				}
			}
			e.Floor += int(e.Dirn)  //Hvis det ikke er kommet noe tegn på at den stopper sier vi at den estimerte heisen sier vi her at den går til en ny etasje
			duration += TRAVEL_TIME //da vil vi også legge til en TRAVEL_TIME kostand for denne opeerasjonen
		}

	}
	return 999 //returnerer høy kostnad dersom heisen er EB_unavailable
}

// func AssignedOrdersAbove(elev elevator.DistributorElevator) bool {
// 	for f := elev.Floor + 1; f < elevator.NumFloors; f++ {
// 		for btn := range elev.AssignedOrders[f] {
// 			if elev.AssignedOrders[f][btn] == elevator.Comfirmed {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func AssignedOrdersBelow(elev elevator.DistributorElevator) bool {
// 	for f := 0; f < elev.Floor; f++ {
// 		for btn := range elev.AssignedOrders[f] {
// 			if elev.AssignedOrders[f][btn] == elevator.Comfirmed {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func requestClearAtCurrentFloor(elev *elevator.DistributorElevator) {
// 	elev.AssignedOrders[elev.Floor][int(elevio.BT_Cab)] = elevator.None
// 	switch {
// 	case elev.Dirn == elevator.Up:
// 		elev.AssignedOrders[elev.Floor][int(elevio.BT_HallUp)] = elevator.None
// 		if !AssignedOrdersAbove(*elev) {
// 			elev.AssignedOrders[elev.Floor][int(elevio.BT_HallDown)] = elevator.None
// 		}
// 	case elev.Dirn == elevator.Down:
// 		elev.AssignedOrders[elev.Floor][int(elevio.BT_HallDown)] = elevator.None
// 		if !AssignedOrdersBelow(*elev) {
// 			elev.AssignedOrders[elev.Floor][int(elevio.BT_HallUp)] = elevator.None
// 		}
// 	}
// }

// func AssignedOrdershouldStop(elev elevator.DistributorElevator) bool {
// 	switch {
// 	case elev.Dirn == elevator.Down:
// 		return elev.AssignedOrders[elev.Floor][int(elevio.BT_HallDown)] == elevator.Comfirmed ||
// 			elev.AssignedOrders[elev.Floor][int(elevio.BT_Cab)] == elevator.Comfirmed ||
// 			!AssignedOrdersBelow(elev)
// 	case elev.Dirn == elevator.Up:
// 		return elev.AssignedOrders[elev.Floor][int(elevio.BT_HallUp)] == elevator.Comfirmed ||
// 			elev.AssignedOrders[elev.Floor][int(elevio.BT_Cab)] == elevator.Comfirmed ||
// 			!AssignedOrdersAbove(elev)
// 	default:
// 		return true
// 	}
// }

// func requestChooseDirnection(elev *elevator.DistributorElevator) {
// 	switch elev.Dirn {
// 	case elevator.Up:
// 		if AssignedOrdersAbove(*elev) {
// 			elev.Dirn = elevator.Up
// 		} else if AssignedOrdersBelow(*elev) {
// 			elev.Dirn = elevator.Down
// 		} else {
// 			elev.Dirn = elevator.Stop
// 		}
// 	case elevator.Down:
// 		fallthrough
// 	case elevator.Stop:
// 		if AssignedOrdersBelow(*elev) {
// 			elev.Dirn = elevator.Down
// 		} else if AssignedOrdersAbove(*elev) {
// 			elev.Dirn = elevator.Up
// 		} else {
// 			elev.Dirn = elevator.Stop
// 		}
// 	}
// }
