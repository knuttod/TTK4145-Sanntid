package cost

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
)

const TRAVEL_TIME = 10
const NumElevators = 4

func Cost(elev *elevator.Elevator, req elevio.ButtonEvent) int {
	if elevator.Behaviour(elev.Behaviour) != elevator.Unavailable {
		e := new(elevator.Elevator)
		*e = *elev //lager en kopi av heisen for å estimere kjøretiden når man legger til den nye orderen
		e.Requests[req.Floor][req.Button] = elevator.Comfirmed

		duration := 0

		switch e.Behaviour {
		case elevator.EB_Idle:
			fsm.ChooseDirection((*e))
			//requestChooseDirnection(e)
			if e.Dirn == elevio.MD_Stop {
				return duration //Dersom IDLE, og hvis det er ingen retning, blir det ingen ekstra kostnad

			}
		case elevator.EB_Moving:
			duration += TRAVEL_TIME / 2 //dersom heisen er i beveglse legger vi til en kostand
			e.Floor += int(e.Dirn)
		case elevator.EB_DoorOpen:
			duration -= elevator.DoorOpenDuration / 2
			//Trekker fra kostnad siden heisen allerede står i ro med dørene åpne og er dermed:
			//Klar til å ta imot nye bestillinger på denne etasjoen, uten ekstra (halvparten) ventetid for å åpne dører

		}
		for {
			if fsm.ShouldStop((*e)) {
				fsm.ClearAtCurrentFloor(*e)
				duration += elevator.DoorOpenDuration
				fsm.ChooseDirection(*e)
				if e.Dirn == elevio.MD_Stop {
					return duration //returner duration når den simulerte heisen har kommet til en stopp
				}
			}
			e.Floor += int(e.Dirn)  //Hvis det ikke er kommet noe tegn på at den stopper sier vi at den estimerte heisen sier vi her at den går til en ny etasje
			duration += TRAVEL_TIME //da vil vi også legge til en TRAVEL_TIME kostand for denne opeerasjonen
		}

	}
	return 999 //returnerer høy kostnad dersom heisen er unavailable
}

// func requestsAbove(elev elevator.DistributorElevator) bool {
// 	for f := elev.Floor + 1; f < elevator.NumFloors; f++ {
// 		for btn := range elev.Requests[f] {
// 			if elev.Requests[f][btn] == elevator.Comfirmed {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func requestsBelow(elev elevator.DistributorElevator) bool {
// 	for f := 0; f < elev.Floor; f++ {
// 		for btn := range elev.Requests[f] {
// 			if elev.Requests[f][btn] == elevator.Comfirmed {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func requestClearAtCurrentFloor(elev *elevator.DistributorElevator) {
// 	elev.Requests[elev.Floor][int(elevio.BT_Cab)] = elevator.None
// 	switch {
// 	case elev.Dirn == elevator.Up:
// 		elev.Requests[elev.Floor][int(elevio.BT_HallUp)] = elevator.None
// 		if !requestsAbove(*elev) {
// 			elev.Requests[elev.Floor][int(elevio.BT_HallDown)] = elevator.None
// 		}
// 	case elev.Dirn == elevator.Down:
// 		elev.Requests[elev.Floor][int(elevio.BT_HallDown)] = elevator.None
// 		if !requestsBelow(*elev) {
// 			elev.Requests[elev.Floor][int(elevio.BT_HallUp)] = elevator.None
// 		}
// 	}
// }

// func requestShouldStop(elev elevator.DistributorElevator) bool {
// 	switch {
// 	case elev.Dirn == elevator.Down:
// 		return elev.Requests[elev.Floor][int(elevio.BT_HallDown)] == elevator.Comfirmed ||
// 			elev.Requests[elev.Floor][int(elevio.BT_Cab)] == elevator.Comfirmed ||
// 			!requestsBelow(elev)
// 	case elev.Dirn == elevator.Up:
// 		return elev.Requests[elev.Floor][int(elevio.BT_HallUp)] == elevator.Comfirmed ||
// 			elev.Requests[elev.Floor][int(elevio.BT_Cab)] == elevator.Comfirmed ||
// 			!requestsAbove(elev)
// 	default:
// 		return true
// 	}
// }

// func requestChooseDirnection(elev *elevator.DistributorElevator) {
// 	switch elev.Dirn {
// 	case elevator.Up:
// 		if requestsAbove(*elev) {
// 			elev.Dirn = elevator.Up
// 		} else if requestsBelow(*elev) {
// 			elev.Dirn = elevator.Down
// 		} else {
// 			elev.Dirn = elevator.Stop
// 		}
// 	case elevator.Down:
// 		fallthrough
// 	case elevator.Stop:
// 		if requestsBelow(*elev) {
// 			elev.Dirn = elevator.Down
// 		} else if requestsAbove(*elev) {
// 			elev.Dirn = elevator.Up
// 		} else {
// 			elev.Dirn = elevator.Stop
// 		}
// 	}
// }
