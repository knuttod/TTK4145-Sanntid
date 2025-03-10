package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"
	"Heis/pkg/timer"
	"fmt"
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
	localAssignedOrder, localRequest chan elevio.ButtonEvent, completedOrderCH chan elevio.ButtonEvent, 
	elevatorStateCh chan msgTypes.ElevatorStateMsg, peerUpdateCh chan peers.PeerUpdate,
	newNodeTx, newNodeRx chan msgTypes.ElevatorStateMsg) {

	var activeLocalOrders [N_floors][N_buttons]bool
	var activeElevators []string
	var new bool = true

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
			if remoteElevatorState.Id != (*e).Id {
				// fmt.Println("External update")
				updateFromRemoteElevator(remoteElevators, e, remoteElevatorState)
				if assignedOrdersKeysCheck(*remoteElevators, *e, activeElevators){
					orderMerger(e, *remoteElevators, activeElevators)
				}
				// fmt.Println("Local: ", (*e).AssignedOrders)
				// fmt.Println("Remote: ", remoteElevatorState.Elevator.AssignedOrders)
				// fmt.Println("own: ", (*e).LocalOrders)

				resetTimer <- timeOutTime
			}
		case remoteElevatorState := <- newNodeRx:
			if remoteElevatorState.Id != (*e).Id {
				if new {
					fmt.Println("New update")
					updateFromRemoteElevator(remoteElevators, e, remoteElevatorState)
					if assignedOrdersKeysCheck(*remoteElevators, *e, activeElevators){
						restartOrdersSynchroniser(e, remoteElevatorState)
						new = false
					}
				}
			}

		case p := <- peerUpdateCh:
			activeElevators = p.Peers
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			// kjøre reassign orders på heisene som ligger i lost. 
			// Antar man kan gjøre noe nice med new for å synkronisere/gi ordre etter avkobling/restart

			if len(p.New) > 0 {
				newNodeTx <- msgTypes.ElevatorStateMsg{Elevator: *e, Id: (*e).Id}
				fmt.Println("newmsg")
			}

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
				if assignedOrdersKeysCheck(*remoteElevators, *e, activeElevators){
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

