package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
)

const TRAVEL_TIME = 10
const NumElevators = 4

func cost(e elevator.Elevator, req elevio.ButtonEvent) int {
	if elevator.ElevatorBehaviour(e.Behaviour) != elevator.EB_Unavailable {

		duration := 0

		switch e.Behaviour {
		case elevator.EB_Idle:
			pair := fsm.ChooseDirection(e)
			e.Dirn = pair.Dirn
			e.Behaviour = pair.Behaviour
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
				e = fsm.ClearAtCurrentFloor(e, nil)
				duration += int(e.Config.DoorOpenDuration_s)
				// might not need this?
				pair := fsm.ChooseDirection(e)
				e.Dirn = pair.Dirn
				e.Behaviour = pair.Behaviour
				// ...
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