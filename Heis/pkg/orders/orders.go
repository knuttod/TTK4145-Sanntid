package orders

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"

	// "Heis/pkg/timer"
	"Heis/pkg/deepcopy"
	// "time"
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


	for {
		ordersToPeersCH <- deepcopy.DeepCopyNettworkElevator(Elevators[selfId])
		select {
		case elev := <- fsmToOrdersCH:
			Elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: *assignedOrders}
		case btn_input := <- buttonPressCH:
			// fmt.Println("assign")
			if assignedOrdersKeysCheck(Elevators, activeElevators){
				assignOrder(assignedOrders, deepcopy.DeepCopyElevatorsMap(Elevators), activeElevators, selfId, btn_input)
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
			}
		case completed_order := <- completedOrderCH:
			fmt.Println("done")
			if (*assignedOrders)[selfId][completed_order.Floor][int(completed_order.Button)] == elevator.Ordr_Confirmed {
				setOrder(assignedOrders, selfId, completed_order.Floor, int(completed_order.Button), elevator.Ordr_Complete)
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
			}

			// select {
		case p := <- peerUpdateCh:
			peerUpdateHandler(assignedOrders, &Elevators, &activeElevators, selfId, p)
			fmt.Println("bef")
			// activeElevatorsCH <- activeElevators 
			fmt.Println("tac")
		case remoteElevatorState := <-remoteElevatorCh: //sender hele tiden
			// fmt.Println("msg")
			if remoteElevatorState.Id != selfId {
				// fmt.Println("remote")
				updateFromRemoteElevator(assignedOrders, &Elevators, remoteElevatorState)
				if assignedOrdersKeysCheck(Elevators, activeElevators){
					// fmt.Println("merge")
					orderMerger(assignedOrders, Elevators, activeElevators, selfId, remoteElevatorState.Id)
					Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: *assignedOrders}
				}
			}
		}

		//kanskje kjøre reassign orders her
		// if assignedOrdersKeysCheck(Elevators, activeElevators) {
		// 	reassignOrders(Elevators, assignedOrders, activeElevators, selfId)
		// }

		for floor := range N_floors {
			for btn := range N_buttons {
				// fmt.Println("Active: ", activeElevators)
				// for _, elev := range activeElevators {
					// fmt.Println(elev, ":", Elevators[elev].AssignedOrders)
				// 	fmt.Println(elev, ":", Elevators[elev].Elevator.Floor)
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
		// duration := time.Since(start)
		// fmt.Println("dur", duration)
	}
}

