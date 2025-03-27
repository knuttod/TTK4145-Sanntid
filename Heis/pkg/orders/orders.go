package orders

import (
	"Heis/pkg/config"
	"Heis/pkg/deepcopy"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/network/network"
	"fmt"
	"log"
)

var (
	numFloors  int
	numBtns    int
	travelTime int
	selfId     string
)

// inits global variables from the config file
func init() {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	numFloors = cfg.NumFloors
	numBtns = cfg.NumBtns
	travelTime = cfg.TravelTime
}

// "Main" function for orders.
// Handles synchronization of orders between elevators by using a cyclic counter.
// Each elevator has a map, assignedOrders, which keeps track of all the orders for all the elevators on the network.
// These maps are compared between the elevators and are updated and acts as a cyclic counter.
// The entries are the orders for the corresponding elevator and there are these orders which are synchronised using a cyclic counter when getting information from another elevator.
// Takes in button presses from FSM and assigns the order to an elevator.
// Takes in completed orders from FSM and marks order as completed in assignedOrders.
// Update on local elevator state is received from FSM and update on remote elevator states from nettwork module.
// Handles reassigning and handling of connection and disconnection of elevators given from the peers update.
// Checks if an order should be started by the local elevator and sends this to the FSM module if it is not busy.
func OrderHandler(id string,
	startLocalOrderCh chan elevio.ButtonEvent, buttonPressCH, completedLocalOrderCH chan elevio.ButtonEvent,
	remoteElevatorUpdateCh chan network.ElevatorStateMsg, peerUpdateCh chan network.PeerUpdate,
	fsmToOrdersCH chan elevator.Elevator, ordersToPeersCH chan elevator.NetworkElevator) {

	//sets global variable for module
	selfId = id

	assignedOrders := assignedOrdersInit(selfId)
	elev := <-fsmToOrdersCH

	elevators := map[string]elevator.NetworkElevator{}
	elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: assignedOrders}

	activeElevators := []string{selfId}
	activeHallLights := initHallLights()

	for {
		ordersToPeersCH <- deepcopy.DeepCopyNettworkElevator(elevators[selfId])
		select {

		// Updates state of this elevator from fsm
		case elev := <-fsmToOrdersCH:
			elevators[selfId] = elevator.NetworkElevator{Elevator: elev, AssignedOrders: assignedOrders}

		// Assigns order for cab or hall button press, forwarded from fsm
		case btnInput := <-buttonPressCH:
			if assignedOrdersKeysCheck(elevators, activeElevators, selfId) {
				assignedOrders = assignOrder(assignedOrders, deepcopy.DeepCopyElevatorsMap(elevators), activeElevators, selfId, btnInput)
				elevators[selfId] = elevator.NetworkElevator{Elevator: elevators[selfId].Elevator, AssignedOrders: assignedOrders}
			}

		// Mark order as completed
		case completedOrder := <-completedLocalOrderCH:
			if assignedOrders[selfId][completedOrder.Floor][int(completedOrder.Button)] == elevator.Ordr_Confirmed {
				assignedOrders[selfId] = setOrder(assignedOrders[selfId], completedOrder.Floor, int(completedOrder.Button), elevator.Ordr_Complete)
				elevators[selfId] = elevator.NetworkElevator{Elevator: elevators[selfId].Elevator, AssignedOrders: assignedOrders}
			}

		// Handles when elevators connects and disconnects from network
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			activeElevators = p.Peers
			peerUpdateHandler(assignedOrders, elevators, activeElevators, selfId, p)
			elevators[selfId] = elevator.NetworkElevator{Elevator: elevators[selfId].Elevator, AssignedOrders: assignedOrders}

		// Updates order and elevator information from other elevators on network.
		case remoteElevatorState := <-remoteElevatorUpdateCh:
			// Updates from itself are ignored, but keeps the select case from stalling
			if remoteElevatorState.Id != selfId {
				assignedOrders, elevators = updateFromRemoteElevator(assignedOrders, elevators, remoteElevatorState)

				if assignedOrdersKeysCheck(elevators, activeElevators, selfId) {
					// merges orders according to cyclic counter
					assignedOrders = orderMerger(assignedOrders, elevators, activeElevators, remoteElevatorState.Id)

					// reassign orders if remote elevator have been obstructed or gotten a motorstop
					assignedOrders = reassignOrdersFromUnavailable(assignedOrders, deepcopy.DeepCopyElevatorsMap(elevators), activeElevators, selfId)
					elevators[selfId] = elevator.NetworkElevator{Elevator: elevators[selfId].Elevator, AssignedOrders: assignedOrders}
				}
			}
		}

		for floor := range numFloors {
			for btn := range numBtns {
				// resets cyclic counter if its only elevator on network
				if len(activeElevators) == 1 {
					if (activeElevators[0] == selfId) && (assignedOrders[selfId][floor][btn] == elevator.Ordr_Complete) {
						assignedOrders[selfId] = setOrder(assignedOrders[selfId], floor, btn, elevator.Ordr_None)
					}
				}

				// sends assigned orders to fsm
				if assignedOrdersKeysCheck(elevators, activeElevators, selfId) {
					assignedOrders = confirmAndStartLocalOrder(assignedOrders, elevators, activeElevators, floor, btn, startLocalOrderCh)
				}
			}
		}

		// hall lights
		if assignedOrdersKeysCheck(elevators, activeElevators, selfId) {
			activeHallLights = setHallLights(assignedOrders, activeElevators, activeHallLights)
		}
	}
}
