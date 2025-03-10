package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	// "fmt"
)

const TRAVEL_TIME = 10
const NumElevators = 4

func cost(e elevator.Elevator, req elevio.ButtonEvent) int {
	if elevator.ElevatorBehaviour(e.Behaviour) != elevator.EB_Unavailable {

		duration := 0
		elev := e

		switch elev.Behaviour {
		case elevator.EB_Idle:
			pair := fsm.ChooseDirection(elev)
			elev.Dirn = pair.Dirn
			elev.Behaviour = pair.Behaviour
			if elev.Dirn == elevio.MD_Stop {
				return duration //Dersom EB_IDLE, og hvis det er ingen retning, blir det ingen ekstra kostnad

			}
		case elevator.EB_Moving:
			duration += TRAVEL_TIME / 2 //dersom heisen er i beveglse legger vi til en kostand
			elev.Floor += int(elev.Dirn)
		case elevator.EB_DoorOpen:
			duration -= int(elev.Config.DoorOpenDuration_s) / 2
			//Trekker fra kostnad siden heisen allerede står i ro med dørene åpne og er dermed:
			//Klar til å ta imot nye bestillinger på denne etasjoen, uten ekstra (halvparten) ventetid for å åpne dører

		}
		for duration < 999 {
			if fsm.ShouldStop(elev) {
				elev = costClearAtCurrentFloor(elev)
				duration += int(elev.Config.DoorOpenDuration_s)
				// might not need this?
				pair := fsm.ChooseDirection(elev)
				elev.Dirn = pair.Dirn
				elev.Behaviour = pair.Behaviour
				// ...
				if elev.Dirn == elevio.MD_Stop {
					return duration //returner duration når den simulerte heisen har kommet til en stopp
				}
			}
			elev.Floor += int(elev.Dirn)  //Hvis det ikke er kommet noe tegn på at den stopper sier vi at den estimerte heisen sier vi her at den går til en ny etasje
			duration += TRAVEL_TIME //da vil vi også legge til en TRAVEL_TIME kostand for denne opeerasjonen
		}
		return 999

	}
	return 999 //returnerer høy kostnad dersom heisen er EB_unavailable
}



//version without sending on completed channel to orders
func costClearAtCurrentFloor(elev elevator.Elevator) elevator.Elevator {
	switch elev.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		for btn := 0; btn < N_buttons; btn++ {
			elev.LocalOrders[elev.Floor][btn] = false
		}
	case elevator.CV_InDirn:
		elev.LocalOrders[elev.Floor][elevio.BT_Cab] = false
		switch elev.Dirn {
		case elevio.MD_Up:
			if (!fsm.Above(elev)) && (elev.LocalOrders[elev.Floor][elevio.BT_HallUp] == false) {
				elev.LocalOrders[elev.Floor][elevio.BT_HallDown] = false
			}
			elev.LocalOrders[elev.Floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if (!fsm.Below(elev)) && (elev.LocalOrders[elev.Floor][elevio.BT_HallDown] == false) {
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
