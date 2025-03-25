package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"

	//"Heis/pkg/message"
	"Heis/pkg/msgTypes"
	//"Heis/pkg/network/peers"
	// "Heis/pkg/orders"
	"log"
	// "fmt"
)

const N_floors = 4
const N_buttons = 3

func Fsm(elev *elevator.Elevator, remoteElevators *map[string]elevator.Elevator, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool, drv_doorTimerStart chan float64, drv_doorTimerFinished chan bool, Tx chan msgTypes.UdpMsg, Rx chan msgTypes.UdpMsg, peerTxEnable chan bool, elevatorStateCh chan msgTypes.ElevatorStateMsg, id string, localAssignedOrder, localRequest chan elevio.ButtonEvent, stateUpdated chan bool) {

	// remoteElevators := make(map[string]elevator.Elevator)

	//Kanskje ikke så robust, uten bruk av channelen
	// Får også problemer med synkronisering med denne
	if elevio.GetFloor() == -1 {
		initBetweenFloors(elev)
	}

	for {
		select {
		case button_input := <-drv_buttons:
			localRequest <- button_input

		case Order := <-localAssignedOrder:
			requestButtonPress(elev, Order.Floor, Order.Button, drv_doorTimerStart, Tx, (*elev).Id)

		case current_floor := <-drv_floors:
			floorArrival(elev, current_floor, drv_doorTimerStart, Tx, id)
			// Send clear floor message

			log.Println("drv_floors: %v", current_floor)
		case obstruction := <-drv_obstr:
			if obstruction {
				(*elev).Obstructed = true
			} else {
				(*elev).Obstructed = false
				fmt.Println("test1")
				drv_doorTimerStart <- (*elev).Config.DoorOpenDuration_s
			}

		case <-drv_doorTimerFinished:
			if !elev.Obstructed {
				DoorTimeout(elev, drv_doorTimerStart)
				log.Println("drv_doortimer timed out")
				DoorTimeout(elev, drv_doorTimerStart)
			}

		case elevatorState := <-elevatorStateCh:
			if elevatorState.Id != (*elev).Id {
				(*remoteElevators)[elevatorState.Id] = elevatorState.Elevator
				// fmt.Println("Updated remoteElevators for ID:", elevatorState.Id)
				// Try merging requests
				remoteElevator, exists := (*remoteElevators)[elevatorState.Id]
				if exists {
					//log.Println("Merging requests for elevator:", elevatorState.Id)

					for id, _ := range *remoteElevators {
						_, exists := (*elev).AssignedOrders[id]
						if !exists {
							fmt.Println("Wrong update")
							(*elev).AssignedOrders[id] = remoteElevator.AssignedOrders[elevatorState.Id]
						}
					}

					fmt.Println("local", (*elev).AssignedOrders)
					fmt.Println("remote", remoteElevator.AssignedOrders)

					// mergeRequests(elev, remoteElevator)
					if assignedOrdersCheck(*remoteElevators, *elev) {
						fmt.Println("statue")
						sendStates(elev, remoteElevator, stateUpdated)
					}
				}
			}
			// case msg := <-Rx:
			// if msg.ButtonPressMsg != nil && msg.ButtonPressMsg.Id != id {
			// 	log.Printf("Received remote button press: %+v\n", msg.ButtonPressMsg)
			// 	// requestButtonPress(elevator, msg.ButtonPressMsg.Floor, msg.ButtonPressMsg.Button, drv_doorTimerStart, Tx, id)
			// }
			// if msg.ClearFloorMsg != nil && msg.ClearFloorMsg.Id != id {
			// 	log.Printf("Received remote clear floor: %+v\n", msg.ClearFloorMsg)

			// 	// Funker ikke som den skal den fjerner alt på samme etasje, tar ikke hensyn til rettning for øyeblikket.
			// 	//				ClearFloor(&elevator, msg.ClearFloorMsg.Floor, msg.ClearFloorMsg.Dirn)
			// }
			// if msg.ElevatorStateMsg != nil && msg.ElevatorStateMsg.Id != id {
			// 	log.Printf("Received remote elevator state: %+v\n", msg.ElevatorStateMsg)

			// 	// **Check if the remote elevator had active requests before merging**
			// 	hasRequests := false
			// 	for _, row := range msg.ElevatorStateMsg.Elevator.Requests {
			// 		for _, req := range row {
			// 			if req {
			// 				hasRequests = true
			// 				break
			// 			}
			// 		}
			// 	}
			// if !hasRequests {
			// 	log.Println("Warning: Received an elevator state with no requests!")
			// }

			// remoteElevators[msg.] = msg.ElevatorStateMsg.Elevator

			// // Try merging requests
			// remoteElevator, exists := remoteElevators[msg.ElevatorStateMsg.Id]
			// if exists {
			// 	log.Println("Merging requests for elevator:", msg.ElevatorStateMsg.Id)
			// 	mergeRequests(elevator, remoteElevator)

			// } else {
			// 	log.Println("No remote elevator state found for ID:", msg.ElevatorStateMsg.Id)
			// }
			// }

		}
	}
}
