package orders

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/network/msgTypes"
	"Heis/pkg/network/peers"
	"log"

	// "Heis/pkg/timer"
	"Heis/pkg/deepcopy"
	// "time"
	"fmt"
)

// This module orders, handles all orders, either comming from a local button press or from updates on nettwork.
// All elevators on the nettwork keeps track of the other elevators order in a map called AssignedOrders, where the keys are elevator id's
// and the values are a 2d slice of assigned orders for the corresponding elevator implemented as a cyclic counter.
// The module is responsible for synchronization of orders and assigning orders to the correct elevator.

// define in config
var (
	N_floors   int
	N_buttons  int
	TravelTime int
)

// inits global variables from the config file
func init() {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	N_floors = cfg.N_floors // Preserving your exact naming
	N_buttons = cfg.N_buttons
	TravelTime = cfg.TravelTime
}

// "Main" function for orders. Takes a ButtonEvent from fsm on localRequest channel when a button is pushed
// and sends an ButtonEvent on localAssignedOrder channel if this eleveator should take order
// Updates local assignedOrders from a remoteElevator sent on elevatorStateCh.
// Also checks if an order to be done by this elevator should be started or not
func OrderHandler(selfId string,
	localAssignedOrder chan elevio.ButtonEvent, buttonPressCH, completedOrderCH chan elevio.ButtonEvent,
	remoteElevatorCh chan msgTypes.ElevatorStateMsg, peerUpdateCh chan peers.PeerUpdate, nettworkDisconnectCh chan bool,
	fsmToOrdersCH chan elevator.Elevator, ordersToPeersCH chan elevator.NetworkElevator) {

	assignedOrders := AssignedOrdersInit(selfId)
	elev := <-fsmToOrdersCH

	Elevators := map[string]elevator.NetworkElevator{}
	Elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: assignedOrders}

	activeElevators := []string{selfId}
	activeHallLights := initHallLights()

	for {
		ordersToPeersCH <- deepcopy.DeepCopyNettworkElevator(Elevators[selfId])
		select {
		case elev := <-fsmToOrdersCH:
			Elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: assignedOrders}
		case btn_input := <-buttonPressCH:
			// fmt.Println("assign")
			if assignedOrdersKeysCheck(Elevators, activeElevators, selfId) {
				assignOrder(&assignedOrders, deepcopy.DeepCopyElevatorsMap(Elevators), activeElevators, selfId, btn_input)
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: assignedOrders}
			}
		case completed_order := <-completedOrderCH:
			// fmt.Println("done")
			if assignedOrders[selfId][completed_order.Floor][int(completed_order.Button)] == elevator.Ordr_Confirmed {
				setOrder(&assignedOrders, selfId, completed_order.Floor, int(completed_order.Button), elevator.Ordr_Complete)
				Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: assignedOrders}
			}

		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			activeElevators = p.Peers
			peerUpdateHandler(&assignedOrders, &Elevators, activeElevators, selfId, p)
			Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: assignedOrders}
		case remoteElevatorState := <-remoteElevatorCh: //sender hele tiden
			// fmt.Println("msg")
			if remoteElevatorState.Id != selfId {
				// fmt.Println("remote")
				updateFromRemoteElevator(&assignedOrders, &Elevators, remoteElevatorState)
				if assignedOrdersKeysCheck(Elevators, activeElevators, selfId) {
					// fmt.Println("merge")
					orderMerger(&assignedOrders, Elevators, activeElevators, selfId, remoteElevatorState.Id)

					// reassign orders if remote elevator have been obstructed or gotten a motorstop

					//denne fjerner hall calls fra en heis som kobler seg på igjen
					reassignOrdersFromUnavailable(deepcopy.DeepCopyElevatorsMap(Elevators), &assignedOrders, activeElevators, selfId)
					// reassignOrders(deepcopy.DeepCopyElevatorsMap(Elevators), &assignedOrders, activeElevators, selfId)
					Elevators[selfId] = elevator.NetworkElevator{Elevator: Elevators[selfId].Elevator, AssignedOrders: assignedOrders}
				}
			}
		case <-nettworkDisconnectCh:
			//To make sure the for loops run even when not reciving remoteElevatorState from peers.reciever
			fmt.Println("disconnected")
			activeElevators = []string{selfId}
			// reassignOrdersFromDisconnectedElevators()
		}

		//kanskje kjøre reassign orders her
		// if assignedOrdersKeysCheck(Elevators, activeElevators, selfId) && (len(activeElevators) > 1){
		// 	reassignOrders(Elevators, &assignedOrders, activeElevators, selfId)
		// }

		for floor := range N_floors {
			for btn := range N_buttons {
				// fmt.Println("Active: ", activeElevators)
				// for _, elev := range activeElevators {
				// 	// for elev := range assignedOrders {
				// 	// fmt.Println(elev, ":", Elevators[elev].AssignedOrders)
				// 	// fmt.Println(elev, ":", Elevators[elev].Elevator.Floor)
				// 	// fmt.Println(elev, ":", Elevators[elev].Elevator.Dirn)
				// 	// if Elevators[elev].Elevator.Obstructed {
				// 	// 	fmt.Println(elev, ": obstructed ", Elevators[elev].Elevator.Obstructed)
				// 	// }
				// // 	// fmt.Println(elev, ": motorstop ", Elevators[elev].Elevator.MotorStop)
				// }
				// fmt.Println("Local: ", assignedOrders)
				// fmt.Println("Remote: ", remoteElevatorState.NetworkElevator.AssignedOrders)
				// fmt.Println("own: ", Elevators[selfId].Elevator.LocalOrders)
				if len(activeElevators) == 1 {
					clearOrder(&assignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn)
				}
				if assignedOrdersKeysCheck(Elevators, activeElevators, selfId) {
					confirmAndStartOrder(&assignedOrders, Elevators, activeElevators, selfId, selfId, floor, btn, localAssignedOrder)
				}
			}
		}

		if assignedOrdersKeysCheck(Elevators, activeElevators, selfId) {
			activeHallLights = setHallLights(assignedOrders, activeElevators, activeHallLights)
		}
		// duration := time.Since(start)
		// fmt.Println("dur", duration)
	}
}
