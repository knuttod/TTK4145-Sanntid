package cost

import (
	"Heis/pkg/config"
	"Heis/pkg/elevio"
)

const TRAVEL_TIME = 10
const NumElevators = 4

func Cost(elev *config.DistributorElevator, req elevio.ButtonEvent) int {
	if elev.Behave != config.Unavailable {
		e := new(config.DistributorElevator)
		*e = *elev //lager en kopi av heisen for å estimere kjøretiden når man legger til den nye orderen
		e.Requests[req.Floor][req.Button] = config.Comfirmed

		duration := 0

		switch e.Behave {
		case config.Idle:
			requestChooseDirection(e)
			if e.Dir == config.Stop {
				return duration //Dersom IDLE og hvis det er ingen retning blir det ingen ekstra kostnad

			}
		case config.Moving:
			duration += TRAVEL_TIME / 2 //dersom heisen er i beveglse legger vi til en kostand
			e.Floor += int(e.Dir)
		case config.DoorOpen:
			duration -= config.DoorOpenDuration / 2
			//Trekker fra kostnad siden heisen allerede står i ro med dørene åpne og er dermed:
			//Klar til å ta imot nye bestillinger på denne etasjoen, uten ekstra (halvparten) ventetid for å åpne dører

		}
		for {
			if requestShouldStop(*e) {
				requestClearAtCurrentFloor(e)
				duration += config.DoorOpenDuration
				requestChooseDirection(e)
				if e.Dir == config.Stop {
					return duration //returner duration når den simulerte heisen har kommet til en stopp
				}
			}
			e.Floor += int(e.Dir)   //Hvis det ikke er kommet noe tegn på at den stopper sier vi at den estimerte heisen sier vi her at den går til en ny etasje
			duration += TRAVEL_TIME //da vil vi også legge til en TRAVEL_TIME kostand for denne opeerasjonen
		}

	}
	return 999 //returnerer høy kostnad dersom heisen er unavailable
}

func requestsAbove(elev config.DistributorElevator) bool {
	for f := elev.Floor + 1; f < config.NumFloors; f++ {
		for btn := range elev.Requests[f] {
			if elev.Requests[f][btn] == config.Comfirmed {
				return true
			}
		}
	}
	return false
}

func requestsBelow(elev config.DistributorElevator) bool {
	for f := 0; f < elev.Floor; f++ {
		for btn := range elev.Requests[f] {
			if elev.Requests[f][btn] == config.Comfirmed {
				return true
			}
		}
	}
	return false
}

func requestClearAtCurrentFloor(elev *config.DistributorElevator) {
	elev.Requests[elev.Floor][int(elevio.BT_Cab)] = config.None
	switch {
	case elev.Dir == config.Up:
		elev.Requests[elev.Floor][int(elevio.BT_HallUp)] = config.None
		if !requestsAbove(*elev) {
			elev.Requests[elev.Floor][int(elevio.BT_HallDown)] = config.None
		}
	case elev.Dir == config.Down:
		elev.Requests[elev.Floor][int(elevio.BT_HallDown)] = config.None
		if !requestsBelow(*elev) {
			elev.Requests[elev.Floor][int(elevio.BT_HallUp)] = config.None
		}
	}
}

func requestShouldStop(elev config.DistributorElevator) bool {
	switch {
	case elev.Dir == config.Down:
		return elev.Requests[elev.Floor][int(elevio.BT_HallDown)] == config.Comfirmed ||
			elev.Requests[elev.Floor][int(elevio.BT_Cab)] == config.Comfirmed ||
			!requestsBelow(elev)
	case elev.Dir == config.Up:
		return elev.Requests[elev.Floor][int(elevio.BT_HallUp)] == config.Comfirmed ||
			elev.Requests[elev.Floor][int(elevio.BT_Cab)] == config.Comfirmed ||
			!requestsAbove(elev)
	default:
		return true
	}
}None      RequestState = 0
Order     RequestState = 1
Comfirmed RequestState = 2
Complete  RequestState = 3
)

func requestChooseDirection(elev *config.DistributorElevator) {
	switch elev.Dir {
	case config.Up:
		if requestsAbove(*elev) {
			elev.Dir = config.Up
		} else if requestsBelow(*elev) {
			elev.Dir = config.Down
		} else {
			elev.Dir = config.Stop
		}
	case config.Down:
		fallthrough
	case config.Stop:
		if requestsBelow(*elev) {
			elev.Dir = config.Down
		} else if requestsAbove(*elev) {
			elev.Dir = config.Up
		} else {
			elev.Dir = config.Stop
		}
	}
}
