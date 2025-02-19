package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/network/peers"
	"Heis/pkg/msgTypes"
	"Heis/pkg/elevator"
	"Heis/pkg/message"
	"log"
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

func Fsm(e *elevator.Elevator, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool, drv_doorTimerStart chan float64, drv_doorTimerFinished chan bool, requestTx chan msgTypes.ButtonPressMsg, requestRx chan msgTypes.ButtonPressMsg, peerTxEnable chan bool, peerUpdateCh chan peers.PeerUpdate) {
	// init state machine between floors

	// elevatorStateMsg := msgTypes.ElevatorStateMsg{
	// 	Elevator: &e,
	// 	Id:       id,
	// }

	// fsm_init(&elevator)
	id := (*e).Id

	//Kanskje ikke så robust, uten bruk av channelen
	if elevio.GetFloor() == -1 {
		initBetweenFloors(e)
	}

	for {

		select {
		case button_input := <-drv_buttons:
			// send button press message

			message.TransmitButtonPress(e, button_input.Floor, button_input.Button, requestTx, id)

			requestButtonPress(e, button_input.Floor, button_input.Button, drv_doorTimerStart)
			log.Println("drv_buttons: %v", button_input)
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
				DoorTimeout(e, drv_doorTimerStart)
				log.Println("drv_doortimer timed out")
			}

		case msg := <-requestRx:
			if msg.Id != id { // Ignore messages from itself
				log.Printf("Received remote button press: %+v\n", msg)
			
				//Act as if the button was pressed locally
				requestButtonPress(e, msg.Floor, msg.Button, drv_doorTimerStart)
			//if msg.ElevatorStateMsg != nil && msg.ElevatorStateMsg.Id != id { // Ignore messages from itself
				//log.Printf("Received remote button press: %+v\n", msg.ElevatorStateMsg)
				// if msg.ElevatorStateMsg.Id == "elevator1" {
				// 	setAllLights(msg.ElevatorStateMsg.Elevator)
				// }
				//fmt.Println("Floor,", *msg.ElevatorStateMsg.Elevator.Floor)
			}
		}
	}
}
