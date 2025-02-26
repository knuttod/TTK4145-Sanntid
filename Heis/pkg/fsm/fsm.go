package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/types"
	"fmt"
	"log"
)

// jonas
const N_floors = 4
const N_buttons = 3

func Fsm(elevator *types.Elevator, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool, drv_doorTimerStart chan float64, drv_doorTimerFinished chan bool, Tx chan types.UdpMsg, Rx chan types.UdpMsg, peerTxEnable chan bool, elevatorStateCh chan types.ElevatorStateMsg, id string) {

	remoteElevators := make(map[string]types.Elevator)

	// init state machine between floors
	fsm_init(elevator)

	//Kanskje ikke så robust, uten bruk av channelen
	if elevio.GetFloor() == -1 {
		initBetweenFloors(elevator)
	}

	for {
		fmt.Println("State: ", elevator.Behaviour)

		select {
		case button_input := <-drv_buttons:
			// send button press message

			requestButtonPress(elevator, button_input.Floor, button_input.Button, drv_doorTimerStart, Tx, id)
			log.Println("drv_buttons: %v", button_input)
		case current_floor := <-drv_floors:
			floorArrival(elevator, current_floor, drv_doorTimerStart, Tx, id)
			// Send clear floor message

			log.Println("drv_floors: %v", current_floor)
		case obstruction := <-drv_obstr:
			if obstruction {
				elevator.Obstructed = true
			} else {
				elevator.Obstructed = false
				drv_doorTimerStart <- elevator.Config.DoorOpenDuration_s
			}

		case <-drv_doorTimerFinished:
			if !elevator.Obstructed {
				DoorTimeout(elevator, drv_doorTimerStart)
				log.Println("drv_doortimer timed out")
			}
		case elevatorState := <-elevatorStateCh:
			remoteElevators[elevatorState.Id] = elevatorState.Elevator
			// fmt.Println("Updated remoteElevators for ID:", elevatorState.Id)
			// Try merging requests
			remoteElevator, exists := remoteElevators[elevatorState.Id]
			if exists {
				log.Println("Merging requests for elevator:", elevatorState.Id)
				mergeRequests(elevator, remoteElevator)
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
