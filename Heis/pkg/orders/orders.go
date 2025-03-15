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
func OrderHandler(e elevator.Elevator, assignedOrders *map[string][][]elevator.RequestState, selfId string,
	localAssignedOrder chan elevio.ButtonEvent, buttonPressCH, completedOrderCH chan msgTypes.FsmMsg,
	remoteElevatorCh chan msgTypes.ElevatorStateMsg, peerUpdateCh chan peers.PeerUpdate,
	newNodeTx, newNodeRx chan msgTypes.ElevatorStateMsg, 
	fsmToOrdersCH chan elevator.Elevator, ordersToPeersCH chan elevator.NetworkElevator) {
	
	
	reassignOrderCH := make(chan elevio.ButtonEvent, 100) //veldig jalla løsning


	Elevators := map[string]elevator.NetworkElevator{}

	Elevators[e.Id] = elevator.NetworkElevator{Elevator: e, AssignedOrders: *assignedOrders}



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
		case elev := <- fsmToOrdersCH:
			Elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: *assignedOrders}
		case elevatorUpdate := <- buttonPressCH:
			fmt.Println("assign")

			Elevators[selfId] = elevator.NetworkElevator{Elevator: elevatorUpdate.Elevator, AssignedOrders: *assignedOrders}
			btn_input := elevatorUpdate.Event

			//For å passe på at man ikke endrer på slicen. Slices er tydelighvis by reference, forstår ikke helt, men er det som er feilen
			copy := make([][]bool, len(Elevators[selfId].Elevator.LocalOrders))
			for i := range copy {
				copy[i] = append([]bool(nil), Elevators[selfId].Elevator.LocalOrders[i]...) // Ensure deep copy
			}
			temp := Elevators[selfId]
			temp.Elevator.LocalOrders = copy
			Elevators[selfId] = temp
			assignOrder(assignedOrders, Elevators, activeElevators, selfId, btn_input) //denne endrer på localOrders mapet. Ikke riktig
			Elevators[selfId] = elevator.NetworkElevator{Elevator: elevatorUpdate.Elevator, AssignedOrders: *assignedOrders}
			

			//resetTimer <- timeOutTime
		
		case request := <- reassignOrderCH:

			fmt.Println("reassign")

			copy := make([][]bool, len(Elevators[selfId].Elevator.LocalOrders))
			for i := range copy {
				copy[i] = append([]bool(nil), Elevators[selfId].Elevator.LocalOrders[i]...) // Ensure deep copy
			}
			temp := Elevators[selfId]
			temp.Elevator.LocalOrders = copy
			
			assignOrder(assignedOrders, Elevators, activeElevators, selfId, request) //denne endrer på localOrders mapet. Ikke riktig
			Elevators[selfId] = temp
			

		case elevatorUpdate := <- completedOrderCH:
			fmt.Println("done")
			
			completed_order := elevatorUpdate.Event

			fmt.Println("Button: ", completed_order.Button)
			fmt.Println("Floor: ", completed_order.Floor)

			temp := (*assignedOrders)[selfId]
			temp[completed_order.Floor][completed_order.Button] = elevator.Complete
			(*assignedOrders)[selfId] = temp

			Elevators[selfId] = elevator.NetworkElevator{Elevator: elevatorUpdate.Elevator, AssignedOrders: *assignedOrders}
			resetTimer <- timeOutTime
		
		case remoteElevatorState := <-remoteElevatorCh:
			if remoteElevatorState.Id != selfId {
				// fmt.Println("External update")
				updateFromRemoteElevator(assignedOrders, &Elevators, remoteElevatorState)
				if assignedOrdersKeysCheck(*assignedOrders, Elevators, selfId, activeElevators){
					orderMerger(assignedOrders, Elevators, activeElevators, selfId)
				}
				// fmt.Println("Local: ", (*assignedOrders))
				// fmt.Println("Remote: ", remoteElevatorState.NetworkElevator.AssignedOrders)
				// fmt.Println("own: ", Elevators[selfId].Elevator.LocalOrders)
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
				resetTimer <- timeOutTime
			}
		case remoteElevatorState := <- newNodeRx:
			if remoteElevatorState.Id != selfId {
				if new {
					fmt.Println("New update")
					updateFromRemoteElevator(assignedOrders, &Elevators, remoteElevatorState)
					if assignedOrdersKeysCheck(*assignedOrders, Elevators, selfId, activeElevators){
						restartOrdersSynchroniser(assignedOrders, remoteElevatorState)
						new = false
					}

				}
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
			}

		case p := <- peerUpdateCh:
			activeElevators = p.Peers
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			// kjøre reassign orders på heisene som ligger i lost. 
			// Antar man kan gjøre noe nice med new for å synkronisere/gi ordre etter avkobling/restart

			if len(p.New) > 0 && !new{
				newNodeTx <- msgTypes.ElevatorStateMsg{
					NetworkElevator: Elevators[selfId],
					Id: selfId,
				}
				fmt.Println("newmsg")
			}
			if len(p.Lost) > 0 {
				for _, elev := range p.Lost {
					temp := Elevators[elev]
					temp.Elevator.Behaviour = elevator.EB_Unavailable
					Elevators[elev] = temp
				}

				reassignOrders(Elevators, *assignedOrders, reassignOrderCH)
				for _, elev := range p.Lost {
					temp := (*assignedOrders)[elev]
					for floor := range N_floors {
						for btn := range N_buttons -1 {
							temp[floor][btn] = elevator.Complete
						}
					}
					(*assignedOrders)[elev] = temp
				}
			}

			resetTimer <- timeOutTime

		case <- timerTimeOut:
			fmt.Println("Timer timed out")
		default:
			//to not stall
			
			
		// case for disconnection or timout for elevator to reassign orders
		// case for synchronization after restart/connection to nettwork

		// tror det er lurt å ha en egen meldingstype som signaliserer at meldingen er etter å ha bli koblet på nettet igjen. 
		// Da kan man håndtere dersom man har forskjellige assignedOrders, F.eks dersom en heis har 0 og en har 2 og man vil ha 2 kan man sette 1 i den som er 0, motsatt sette 3 i den som er 2.
		}

		

		// Check if an unstarted assigned order should be started
		for floor := range N_floors {
			for btn := range N_buttons {
				if (*assignedOrders)[selfId][floor][btn] != elevator.Confirmed {
					activeLocalOrders[floor][btn] = false
				}
				// fmt.Println("Active: ", activeElevators)
				// fmt.Println("Local: ", (*assignedOrders))
				// fmt.Println("Remote: ", remoteElevatorState.NetworkElevator.AssignedOrders)
				// fmt.Println("own: ", Elevators[selfId].Elevator.LocalOrders)
				if assignedOrdersKeysCheck(*assignedOrders, Elevators, selfId, activeElevators){
					if len(activeElevators) == 1 {
						confirmOrCloseOrders(assignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn)
					}
					if shouldStartLocalOrder(*assignedOrders, Elevators, activeElevators, selfId, floor, btn) && !activeLocalOrders[floor][btn] {
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

		//might need buffering/can be stalled by the 15ms wait time in peers
		ordersToPeersCH <- Elevators[selfId]
	}
}

