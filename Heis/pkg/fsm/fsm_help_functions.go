package fsm

import (
	"Heis/pkg/elevio"
	"Heis/pkg/types"

	//	"fmt"
	"time"
)

func fsm_init(e *types.Elevator) {
	// initialize the (*e) struct
	(*e).Floor = -1
	(*e).Dirn = elevio.MD_Stop
	(*e).Behaviour = types.EB_Idle
	(*e).Config.ClearRequestVariant = types.CV_InDirn
	(*e).Config.DoorOpenDuration_s = 3.0
	(*e).Requests = make([][]bool, N_floors)
	for i := range (*e).Requests {
		(*e).Requests[i] = make([]bool, N_buttons)
	}
}

func initBetweenFloors(e *types.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	(*e).Dirn = elevio.MD_Down
	(*e).Behaviour = types.EB_Moving
}

func requestButtonPress(e *types.Elevator, btn_floor int, btn_type elevio.ButtonType, drv_doorTimer chan float64, Tx chan types.UdpMsg, id string) {
	//print functions??
	if btn_type != elevio.BT_Cab {
		buttonPressMsg := types.ButtonPressMsg{
			Floor:  btn_floor,
			Button: btn_type,
			Id:     id,
		}

		// Retransmit to reduce redundancy
		for i := 0; i < 10; i++ {
			Tx <- types.UdpMsg{ButtonPressMsg: &buttonPressMsg}
			time.Sleep(10 * time.Millisecond)
		}
	}

	switch (*e).Behaviour {
	case types.EB_DoorOpen:
		if ShouldClearImmediately((*e), btn_floor, btn_type) {
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			//drv_doorTimer <- 0.0
		} else {
			(*e).Requests[btn_floor][btn_type] = true
		}

	case types.EB_Moving:
		(*e).Requests[btn_floor][btn_type] = true

	case types.EB_Idle:
		(*e).Requests[btn_floor][btn_type] = true
		var pair DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case types.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			//drv_doorTimer <- 0.0
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s
			(*e) = ClearAtCurrentFloor((*e))

		case types.EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
			//clear something at this floor??

		case types.EB_Idle:
			//need something here?
		}

	}
	setAllLights(e)
}

func floorArrival(e *types.Elevator, newFloor int, drv_doorTimer chan float64, Tx chan types.UdpMsg, id string) {
	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)

	switch e.Behaviour {
	case types.EB_Moving:
		if ShouldStop(*e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			*e = ClearAtCurrentFloor(*e)
			drv_doorTimer <- e.Config.DoorOpenDuration_s
			setAllLights(e)
			(*e).Behaviour = types.EB_DoorOpen

			// Broadcast clear floor message
			clearFloorMsg := types.ClearFloorMsg{
				Floor: newFloor,
				Dirn:  e.Dirn,
				Id:    id,
			}
			for i := 0; i < 10; i++ {
				Tx <- types.UdpMsg{ClearFloorMsg: &clearFloorMsg}
			}
		}
	}
}

func DoorTimeout(e *types.Elevator, drv_doorTimer chan float64) {

	switch (*e).Behaviour {
	case types.EB_DoorOpen:
		var pair DirnBehaviourPair = chooseDirection((*e))
		(*e).Dirn = pair.Dirn
		(*e).Behaviour = pair.Behaviour

		switch (*e).Behaviour {
		case types.EB_DoorOpen:
			drv_doorTimer <- (*e).Config.DoorOpenDuration_s //????
			//drv_doorTimer <- 0.0
			(*e) = ClearAtCurrentFloor((*e))
			setAllLights(e)

		//lagt inn selv
		case types.EB_Moving:
			elevio.SetMotorDirection((*e).Dirn)
		//

		case types.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection((*e).Dirn)

		}
	}
}

func setAllLights(e *types.Elevator) {
	//set ligths
	for floor := 0; floor < N_floors; floor++ {
		for btn := 0; btn < N_buttons; btn++ {
			if e.Requests[floor][btn] {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			}
		}
	}
}

func clearCabIfOtherCleared(e *types.Elevator, floor int) {
	if e.Requests[floor][elevio.BT_Cab] {
		e.Requests[floor][elevio.BT_Cab] = false
		setAllLights(e) // Update lights
	}
}

func monitorRemoteCabCalls(elevator *types.Elevator, remoteElevators *map[string]types.Elevator) {
	for {
		time.Sleep(100 * time.Millisecond) // Check periodically

		for _, remote := range *remoteElevators {
			for floor := 0; floor < N_floors; floor++ {
				// If local elevator has a cab request, but another elevator has cleared it
				if elevator.Requests[floor][elevio.BT_Cab] && !remote.Requests[floor][elevio.BT_Cab] {
					clearCabIfOtherCleared(elevator, floor)
				}
			}
		}
	}
}

func broadcastElevatoStates(elevator *types.Elevator, id string, Tx chan types.UdpMsg) {
	for {
		time.Sleep(100 * time.Millisecond) // Avoid flooding the network
		elevatorStateMsg := types.ElevatorStateMsg{
			Elevator: *elevator,
			Id:       id,
		}
		Tx <- types.UdpMsg{ElevatorStateMsg: &elevatorStateMsg}
	}
}

func updateRemoteElevators(Rx chan types.UdpMsg, remoteElevators map[string]types.Elevator, localElevator *types.Elevator, id string) {
	for {
		select {
		case msg := <-Rx:
			if msg.ElevatorStateMsg != nil && msg.ElevatorStateMsg.Id != id {
				remoteElevator := msg.ElevatorStateMsg.Elevator

				// Only sync hall orders if the remote elevator is in a state where it's safe to copy them
				if remoteElevator.Behaviour == (types.EB_DoorOpen) {
					syncHallRequests(localElevator, remoteElevator)
				}

				// Update the remote elevator states
				remoteElevators[msg.ElevatorStateMsg.Id] = remoteElevator
			}
		}
	}
}

func syncHallRequests(local *types.Elevator, remote types.Elevator) {
	for floor := 0; floor < N_floors; floor++ {
		for btn := 0; btn < N_buttons; btn++ {
			if btn != int(elevio.BT_Cab) { // Only sync hall calls
				if remote.Requests[floor][btn] {
					local.Requests[floor][btn] = true
				}
			}
		}
	}
}
