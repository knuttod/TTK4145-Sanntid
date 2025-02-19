package message

import (
	"Heis/pkg/elevator"
	"Heis/pkg/msgTypes"
	"Heis/pkg/elevio"
	"time"
)


func TransmitState(e *elevator.Elevator, Tx chan msgTypes.ElevatorStateMsg, id string) {
	elevatorStateMsg := msgTypes.ElevatorStateMsg{
			Elevator: e,
			Id:       id,
		}
	for {
		//Tx <- msgTypes.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
		Tx <- elevatorStateMsg
		time.Sleep(10 * time.Millisecond)
	}
}

func TransmitButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType, Tx chan msgTypes.ButtonPressMsg, id string) {
	buttonPressMsg := msgTypes.ButtonPressMsg{
		Floor:  btn_floor,
		Button: btn_type,
		Id:     id,
	}

	// Retransmit to reduce redundancy
	for i := 0; i < 30; i++ {
		Tx <- buttonPressMsg
		time.Sleep(10 * time.Millisecond)
	}
}