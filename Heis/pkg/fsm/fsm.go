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

// type ElevatorBehaviour int

// const (
// 	EB_Idle ElevatorBehaviour = iota
// 	EB_DoorOpen
// 	EB_Moving
// )

// type ClearRequestVariant int

// const (
// 	CV_ALL ClearRequestVariant = iota
// 	CV_InDirn
// )

// type Elevator struct {
// 	Floor      int
// 	Dirn       elevio.MotorDirection
// 	Requests   [][]bool
// 	Behaviour  ElevatorBehaviour
// 	Obstructed bool

// 	Config struct { //type?
// 		ClearRequestVariant ClearRequestVariant
// 		DoorOpenDuration_s  float64
// 	}
// }

func Fsm(e *elevator.Elevator, order chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool, drv_doorTimerStart chan float64, drv_doorTimerFinished chan bool, requestTx chan msgTypes.ButtonPressMsg, requestRx chan msgTypes.ButtonPressMsg, peerTxEnable chan bool, peerUpdateCh chan peers.PeerUpdate) {
	// init state machine between floors

	// elevatorStateMsg := msgTypes.ElevatorStateMsg{
	// 	Elevator: &e,
	// 	Id:       id,
	// }

	// fsm_init(&elevator)

	//Kanskje ikke så robust, uten bruk av channelen
	if elevio.GetFloor() == -1 {
		initBetweenFloors(e)
	}

	for {
		select {
		case button_input := <-order:
			// send button press message

			// send unconfirmed message to OrderMerger
			// When order is confirmed and is assigned from cost function calculations to this elevator send to request button press

			// Need a way to tell others order is done

			// message.TransmitButtonPress(e, button_input.Floor, button_input.Button, requestTx, (*e).Id)

			log.Println("drv_buttons: %v", button_input)
			fmt.Println("Button input: ", button_input)
			requestButtonPress(e, button_input.Floor, button_input.Button, drv_doorTimerStart)
		case current_floor := <-drv_floors:
			floorArrival(e, current_floor, drv_doorTimerStart)
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
			if !(*e).Obstructed {
				log.Println("drv_doortimer timed out")
				DoorTimeout(e, drv_doorTimerStart)
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
