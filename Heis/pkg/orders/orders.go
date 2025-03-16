package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"
	// "Heis/pkg/timer"
	"Heis/pkg/deepcopy"
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
func OrderHandler(e elevator.Elevator, assignedOrders *map[string][][]elevator.OrderState, selfId string,
	localAssignedOrder chan elevio.ButtonEvent, buttonPressCH, completedOrderCH chan elevio.ButtonEvent,
	remoteElevatorCh chan msgTypes.ElevatorStateMsg, peerUpdateCh chan peers.PeerUpdate,
	newNodeTx, newNodeRx chan msgTypes.ElevatorStateMsg, 
	fsmToOrdersCH chan elevator.Elevator, ordersToPeersCH chan elevator.NetworkElevator) {
	


	Elevators := map[string]elevator.NetworkElevator{}
	Elevators[e.Id] = elevator.NetworkElevator{Elevator: e, AssignedOrders: *assignedOrders}

	var activeElevators []string

	
	// resetTimer := make(chan float64)
	// timerTimeOut := make(chan bool)
	// timeOutTime := 10.0
	// go timer.Timer(resetTimer, timerTimeOut)
	//Lage en watchdog funksjon som gjør noe dersom timeren utløper
	// sette opp en timer som må kontinuerlig resettes, dersom den ikke gjør det send på en channel her som da gjør et eller annet, f.eks resetter programmet
	// dette kan brukes for å sjekke for motorstopp?
	// kan være en ide å resette heisen dersom dette skjer


	// denne blir stuck noen ganger, vet ikke helt hvorfor??
	for {
		// ordersToPeersCH <- Elevators[selfId]
		ordersToPeersCH <- deepcopy.DeepCopyNettworkElevator(Elevators[selfId])
		
		select {
		case elev := <- fsmToOrdersCH:
			Elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: *assignedOrders}
		default:
			//non blocking
		}

		select {
		case btn_input := <- buttonPressCH:
			fmt.Println("assign")
			assignOrder(assignedOrders, deepcopy.DeepCopyElevatorsMap(Elevators), activeElevators, selfId, btn_input)
			Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
			
			//starte timer for å sjekke tid

		
		//denne trenger kun å si ifra at den er ferdig for timer
		case completed_order := <- completedOrderCH:

			fmt.Println("done")

			if (*assignedOrders)[selfId][completed_order.Floor][int(completed_order.Button)] == elevator.Ordr_Confirmed {
				setOrder(assignedOrders, selfId, completed_order.Floor, int(completed_order.Button), elevator.Ordr_Complete)
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
			}

			// //si til timer at den ble fullført innen tiden
			// resetTimer <- timeOutTime
		
		case remoteElevatorState := <-remoteElevatorCh:
			if remoteElevatorState.Id != selfId {
				// fmt.Println("remote")
				updateFromRemoteElevator(assignedOrders, &Elevators, remoteElevatorState)
				if assignedOrdersKeysCheck(Elevators, activeElevators){
					// fmt.Println("merge")
					orderMerger(assignedOrders, Elevators, activeElevators, selfId, remoteElevatorState.Id)
					Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
				}
				
				// resetTimer <- timeOutTime
			}

		case p := <- peerUpdateCh:
			activeElevators = p.Peers
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			// kjøre reassign orders på heisene som ligger i lost. 
			// Antar man kan gjøre noe nice med new for å synkronisere/gi ordre etter avkobling/restart

			fmt.Println("len", len(p.Lost))
			if len(p.Lost) > 0 {
				for _, elev := range p.Lost {
					temp := Elevators[elev]
					temp.Elevator.Behaviour = elevator.EB_Unavailable
					Elevators[elev] = temp
				}
				fmt.Println("lost")

				//alt dette bør bli en funksjon og ikke kjøres her
				reassignOrders(deepcopy.DeepCopyElevatorsMap(Elevators), assignedOrders, activeElevators, selfId)
				// for _, elev := range p.Lost {
				// 	temp := Elevators[elev]
				// 	tempOrders := temp.AssignedOrders
				// 	for floor := range N_floors {
				// 		for btn := range N_buttons -1 {
				// 			setOrder(&tempOrders, elev, floor, btn, elevator.Ordr_None)
				// 			temp.AssignedOrders = tempOrders
				// 			Elevators[elev] = temp
				// 		}
				// 	}
				// }
			}

			//sette hall orders på seg selv til unkown dersom man ikke har noen andre peers

			if len(p.Peers) == 1 {
				for id := range (*assignedOrders) {
					// Kanskje ikke sette sine egne til unkown
					if id == selfId {
						continue
					}
					for floor := range N_floors {
						for btn := range (N_buttons - 1) {
							setOrder(assignedOrders, id, floor, btn, elevator.Ordr_Unknown)
						}
					}
				}
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
			}
	

			// resetTimer <- timeOutTime

		// case <- timerTimeOut:
			// fmt.Println("Timer timed out")
		// default:
			//to not stall
		}

		//trenger å kjøre denne her for motorstopp senere
		// reassignOrders(deepcopy.DeepCopyElevatorsMap(Elevators), assignedOrders, activeElevators, selfId)

		// Check if an unstarted assigned order should be started
		for floor := range N_floors {
			for btn := range N_buttons {
				// fmt.Println("Active: ", activeElevators)
				// for _, elev := range activeElevators {
				// 	fmt.Println(elev, ":", Elevators[elev].AssignedOrders)
				// }
				// fmt.Println("Local: ", (*assignedOrders))
				// fmt.Println("Remote: ", remoteElevatorState.NetworkElevator.AssignedOrders)
				// fmt.Println("own: ", Elevators[selfId].Elevator.LocalOrders)
				if len(activeElevators) == 1 {
					clearOrder(assignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn)
				}
				if assignedOrdersKeysCheck(Elevators, activeElevators){
					confirmAndStartOrder(assignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn, localAssignedOrder)
				}
			}
		}
		//non blocking

		//set lights
		//litt buggy på macsimmen, kanskje bedre andre steder?
		// setAllHallLightsfromRemote(*remoteElevators, activeElevators, (*e).Id)

	}
}

