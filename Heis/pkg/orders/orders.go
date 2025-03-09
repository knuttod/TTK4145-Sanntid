package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"
	"Heis/pkg/timer"
	"fmt"
	"reflect"
	"sort"
)

// This module orders, handles all orders, either comming from a local button press or from updates on nettwork.
// All elevators on the nettwork keeps track of the other elevators order in a map called AssignedOrders, where the keys are elevator id's
// and the values are a 2d slice of assigned orders for the corresponding elevator implemented as a cyclic counter.
// The module is responsible for synchronization of orders and assigning orders to the correct elevator.

//temporarly, should be loaded from config
const N_floors = 4
const N_buttons = 3


// "Main" function for orders. Takes a ButtonEvent from fsm on localRequest channel when a button is pushed
// and sends an ButtonEvent on localAssignedOrder channel if this eleveator should take order
// Updates local assignedOrders from a remoteElevator sent on elevatorStateCh.
// Also checks if an order to be done by this elevator should be started or not
func OrderHandler(e *elevator.Elevator, remoteElevators *map[string]elevator.Elevator, 
	localAssignedOrder, localRequest chan elevio.ButtonEvent, elevatorStateCh chan msgTypes.ElevatorStateMsg, completedOrderCH chan elevio.ButtonEvent, peerUpdateCh chan peers.PeerUpdate) {
	var activeLocalOrders [N_floors][N_buttons]bool
	var activeElevators []string

	

	resetTimer := make(chan float64)
	timerTimeOut := make(chan bool)
	timeOutTime := 10.0
	go timer.Timer(resetTimer, timerTimeOut)
	//Lage en watchdog funksjon som gjør noe dersom timeren utløper
	// sette opp en timer som må kontinuerlig resettes, dersom den ikke gjør det send på en channel her som da gjør et eller annet, f.eks resetter programmet
	// dette kan brukes for å sjekke for motorstopp?
	// kan være en ide å resette heisen dersom dette skjer


	// denne blir stuck noen ganger, vet ikke helt hvorfor??
	for {
		select {
		case btn_input := <- localRequest:
			
			//For å passe på at man ikke endrer på slicen. Slices er tydelighvis by reference, forstår ikke helt, men er det som er feilen
			local := make([][]bool, len((*e).LocalOrders))
			for i := range local {
				local[i] = append([]bool(nil), (*e).LocalOrders[i]...) // Ensure deep copy
			}
			fmt.Println("assign")
			assignOrder(e, *remoteElevators, activeElevators, btn_input) //denne endrer på localOrders mapet. Ikke riktig
			(*e).LocalOrders = local

			resetTimer <- timeOutTime

		case completed_order := <- completedOrderCH:
			fmt.Println("done")
			temp := (*e).AssignedOrders[(*e).Id]
			temp[completed_order.Floor][completed_order.Button] = elevator.Complete
			(*e).AssignedOrders[(*e).Id] = temp

			resetTimer <- timeOutTime
		
		case remoteElevatorState := <-elevatorStateCh:
			fmt.Println("Update")
			if remoteElevatorState.Id != (*e).Id {
				fmt.Println("External update")
				updateFromRemoteElevator(remoteElevators, e, remoteElevatorState)
				if assignedOrdersCheck(*remoteElevators, *e, activeElevators){
					orderMerger(e, *remoteElevators, activeElevators)
				}
				// fmt.Println("Local: ", (*e).AssignedOrders)
				// fmt.Println("Remote: ", remoteElevatorState.Elevator.AssignedOrders)
				// fmt.Println("own: ", (*e).LocalOrders)

				resetTimer <- timeOutTime
			}
		case p := <- peerUpdateCh:
			activeElevators = p.Peers
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			// kjøre reassign orders på heisene som ligger i lost. 
			// Antar man kan gjøre noe nice med new for å synkronisere/gi ordre etter avkobling/restart

			resetTimer <- timeOutTime

		case <- timerTimeOut:
			fmt.Println("Timer timed out")
			
			
		// case for disconnection or timout for elevator to reassign orders
		// case for synchronization after restart/connection to nettwork

		// tror det er lurt å ha en egen meldingstype som signaliserer at meldingen er etter å ha bli koblet på nettet igjen. 
		// Da kan man håndtere dersom man har forskjellige assignedOrders, F.eks dersom en heis har 0 og en har 2 og man vil ha 2 kan man sette 1 i den som er 0, motsatt sette 3 i den som er 2.
		}

		// Check if an unstarted assigned order should be started
		for floor := range N_floors {
			for btn := range N_buttons {
				if (*e).AssignedOrders[(*e).Id][floor][btn] != elevator.Confirmed {
					activeLocalOrders[floor][btn] = false
				}
				if assignedOrdersCheck(*remoteElevators, *e, activeElevators){
					if shouldStartLocalOrder(e, *remoteElevators, activeElevators, (*e).Id, floor, btn) && !activeLocalOrders[floor][btn] {
						// fmt.Println("her")
						localAssignedOrder <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.ButtonType(btn),
						}
						activeLocalOrders[floor][btn] = true
					}
				}
			}
		}
		//set lights
		//litt buggy på macsimmen, kanskje bedre andre steder?
		// setAllHallLightsfromRemote(*remoteElevators, activeElevators, (*e).Id)
	}
}

// Updates state in assignedOrders for local elevator if it is behind remoteElevators
func orderMerger(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string) {
	var currentState elevator.RequestState
	var updateState elevator.RequestState

	for id, _ := range (*e).AssignedOrders {	
		for floor := range N_floors{
			for btn := range N_buttons {
				// for _, elev := range remoteElevators {
				// 	currentState = (*e).AssignedOrders[id][floor][btn]
				// 	updateState = elev.AssignedOrders[id][floor][btn]

				// 	if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
				// 		temp := (*e).AssignedOrders[id]
				// 		temp[floor][btn] = updateState
				// 		(*e).AssignedOrders[id] = temp
				// 	}
				// }
				for _, elev := range activeElevators {
					if elev == (*e).Id {
						continue
					}
					currentState = (*e).AssignedOrders[id][floor][btn]
					updateState = remoteElevators[elev].AssignedOrders[id][floor][btn]

					if updateState - currentState == 1 || (updateState == elevator.None && currentState == elevator.Complete) {
						temp := (*e).AssignedOrders[id]
						temp[floor][btn] = updateState
						(*e).AssignedOrders[id] = temp
					}
				}	
				confirmOrCloseOrders(e, remoteElevators, activeElevators, id, floor, btn)
			}
		}
	}
}

// Check if all nodes on nettwork are synced on certain order
func ordersSynced(e elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, id string, floor, btn int) bool {
	// for _, elev := range remoteElevators {
	// 	if (*e).AssignedOrders[id][floor][btn] != elev.AssignedOrders[id][floor][btn] {
	// 		return false
	// 	}
	// }
	// return true
	for _, elev := range activeElevators {
		if elev == (e).Id {
			continue
		}
		if (e).AssignedOrders[id][floor][btn] != remoteElevators[elev].AssignedOrders[id][floor][btn] {
			return false
		}
	}
	return true
}


// Increments cyclic counter if all nodes on nettwork are synced on order and state is order or completed
func confirmOrCloseOrders(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, id string, floor, btn int) {
	if ordersSynced(*e, remoteElevators, activeElevators, id, floor, btn) {
		if  (*e).AssignedOrders[id][floor][btn] == elevator.Order {
			temp := (*e).AssignedOrders[id]
			temp[floor][btn] = elevator.Confirmed
			(*e).AssignedOrders[id] = temp
		}
		if  (*e).AssignedOrders[id][floor][btn] == elevator.Complete{
			temp := (*e).AssignedOrders[id]
			temp[floor][btn] = elevator.None
			(*e).AssignedOrders[id] = temp
		}
	}
}

// Should start order when elevators are synced and order is confirmed
func shouldStartLocalOrder(e *elevator.Elevator, remoteElevators map[string]elevator.Elevator, activeElevators []string, id string, floor, btn int) bool {
	return ordersSynced(*e, remoteElevators, activeElevators, id, floor, btn) && (*e).AssignedOrders[id][floor][btn] == elevator.Confirmed
}

// Updates remoteElevator map and adds entries for remote Elevator in assignedOrders map for local elevator if they do not exist yet. 
func updateFromRemoteElevator(remoteElevators * map[string]elevator.Elevator, e * elevator.Elevator, remoteElevatorState msgTypes.ElevatorStateMsg) {
	remoteElevator := remoteElevatorState.Elevator
	(*remoteElevators)[remoteElevatorState.Id] = remoteElevator
	
	for id, _ := range *remoteElevators {
		_, exists := (*e).AssignedOrders[id]
		if !exists {
			(*e).AssignedOrders[id] = remoteElevator.AssignedOrders[remoteElevatorState.Id]
		}
	}
}

// Checks if all active elevators in remoteElevators have an assignedOrders map with keys for all active elevators on nettwork 
func assignedOrdersCheck(remoteElevators map[string]elevator.Elevator, e elevator.Elevator, activeElevators []string) bool {
	
	var localKeys []string
	var remoteKeys []string
	
	for k, _ := range e.AssignedOrders{
		localKeys = append(localKeys, k)
	}
	
	// for _, elev := range remoteElevators {
	// 	remoteKeys = []string{}
	// 	for k, _ := range elev.AssignedOrders{
	// 		remoteKeys = append(remoteKeys, k)
	// 	}
	// 	sort.Strings(localKeys)
	// 	sort.Strings(remoteKeys)
	// 	// fmt.Println("local keys", localKeys)
	// 	// fmt.Println("External keys", remoteKeys)

	// 	if !reflect.DeepEqual(localKeys, remoteKeys) {
	// 		return false
	// 	}
	// }
	for _, elev := range activeElevators {
		if elev == e.Id {
			continue
		}
		remoteKeys = []string{}
		for k, _ := range remoteElevators[elev].AssignedOrders{
			remoteKeys = append(remoteKeys, k)
		}
		sort.Strings(localKeys)
		sort.Strings(remoteKeys)
		// fmt.Println("local keys", localKeys)
		// fmt.Println("External keys", remoteKeys)

		if !reflect.DeepEqual(localKeys, remoteKeys) {
			return false
		}
	}
	return true
}

func setAllHallLightsfromRemote(remoteElevators map[string]elevator.Elevator, activeElevators []string, selfId string) {
	
	var setLight bool
	for floor := 0; floor < N_floors; floor++ {
		for btn := 0; btn < N_buttons; btn++ {
			setLight = false
			for _, elev := range activeElevators {
				if elev == selfId {
					continue
				}
				if remoteElevators[elev].AssignedOrders[elev][floor][btn] == elevator.Confirmed{
					setLight = true	
				}
			}
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, setLight)
		}
	}
}