package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	//"Heis/pkg/message"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"
	"log"
	"fmt"
)

// jonas
const N_floors = 4
const N_buttons = 3

func Fsm(elevator *elevator.Elevator, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool, drv_doorTimerStart chan float64, drv_doorTimerFinished chan bool, Tx chan types.UdpMsg, Rx chan types.UdpMsg, peerTxEnable chan bool, peerUpdateCh chan peers.PeerUpdate, id string) {

	remoteElevators := make(map[string]elevator.Elevator)

	// init state machine between floors
	fsm_init(elevator)

	//Kanskje ikke så robust, uten bruk av channelen
	if elevio.GetFloor() == -1 {
		initBetweenFloors(elevator)
	}

	for {
		select {
		case button_input := <-drv_buttons:
			//assign
			if ordersSynced(e, remoteElevators, button_input.floor, button_input.btn){
				AssignOrder(remoteElevators, (*e).Id, button_input)
			}
			// send button press message

			// send unconfirmed message to OrderMerger
			// When order is confirmed and is assigned from cost function calculations to this elevator send to request button press

			// Need a way to tell others order is done

			// message.TransmitButtonPress(e, button_input.Floor, button_input.Button, requestTx, (*e).Id)

			log.Println("drv_buttons: %v", button_input)
			fmt.Println("Button input: ", button_input)
			// requestButtonPress(e, button_input.Floor, button_input.Button, drv_doorTimerStart)
		case Order := <- startOrder:
			requestButtonPress(e, startOrder.Floor, startOrder.Button, drv_doorTimerStart)
		
		case current_floor := <-drv_floors:
			floorArrival(elevator, current_floor, drv_doorTimerStart, Tx, id)
			// Send clear floor message

			log.Println("drv_floors: %v", current_floor)
		case obstruction := <-drv_obstr:
			if obstruction {
				(*e).Obstructed = true
			} else {
				(*e).Obstructed = false
				drv_doorTimerStart <- (*e).Config.DoorOpenDuration_s
			}

		case <-drv_doorTimerFinished:
			if !elevator.Obstructed {
				DoorTimeout(elevator, drv_doorTimerStart)
				log.Println("drv_doortimer timed out")
				DoorTimeout(e, drv_doorTimerStart)
			}

		case msg := <-Rx:
			// if msg.ButtonPressMsg != nil && msg.ButtonPressMsg.Id != id {
			// 	log.Printf("Received remote button press: %+v\n", msg.ButtonPressMsg)
			// 	// requestButtonPress(elevator, msg.ButtonPressMsg.Floor, msg.ButtonPressMsg.Button, drv_doorTimerStart, Tx, id)
			// }
			// if msg.ClearFloorMsg != nil && msg.ClearFloorMsg.Id != id {
			// 	log.Printf("Received remote clear floor: %+v\n", msg.ClearFloorMsg)

			// 	// Funker ikke som den skal den fjerner alt på samme etasje, tar ikke hensyn til rettning for øyeblikket.
			// 	//				ClearFloor(&elevator, msg.ClearFloorMsg.Floor, msg.ClearFloorMsg.Dirn)
			// }
			if msg.ElevatorStateMsg != nil && msg.ElevatorStateMsg.Id != id {
				log.Printf("Received remote elevator state: %+v\n", msg.ElevatorStateMsg)

				// **Check if the remote elevator had active requests before merging**
				hasRequests := false
				for _, row := range msg.ElevatorStateMsg.Elevator.Requests {
					for _, req := range row {
						if req {
							hasRequests = true
							break
						}
					}
				}
				if !hasRequests {
					log.Println("Warning: Received an elevator state with no requests!")
				}

				remoteElevators[msg.ElevatorStateMsg.Id] = msg.ElevatorStateMsg.Elevator

				// Try merging requests
				remoteElevator, exists := remoteElevators[msg.ElevatorStateMsg.Id]
				if exists {
					log.Println("Merging requests for elevator:", msg.ElevatorStateMsg.Id)
					mergeRequests(&elevator.Requests, remoteElevator.Requests)

				} else {
					log.Println("No remote elevator state found for ID:", msg.ElevatorStateMsg.Id)
				}
			}

		case msg := <-requestRx:
			if msg.Id != (*e).Id { // Ignore messages from itself
				log.Printf("Received remote button press: %+v\n", msg)
				SetAllLightsOrder((*e).GlobalOrders, e)
				//Act as if the button was pressed locally
				//requestButtonPress(e, msg.Floor, msg.Button, drv_doorTimerStart)

			} 
			
				
			//if msg.ElevatorStateMsg != nil && msg.ElevatorStateMsg.Id != id { // Ignore messages from itself
				//log.Printf("Received remote button press: %+v\n", msg.ElevatorStateMsg)
				// if msg.ElevatorStateMsg.Id == "elevator1" {
				// 	setAllLights(msg.ElevatorStateMsg.Elevator)
				// }
				//fmt.Println("Floor,", *msg.ElevatorStateMsg.Elevator.Floor)
			
		}
	}
}
