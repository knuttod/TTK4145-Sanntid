package fsm

import (
	"Heis/pkg/deepcopy"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/timer"
	"fmt"
	// "fmt"
)

// define in config
const N_floors = 4
const N_buttons = 3

// FSM handles core logic of a single Elevator. Interacts with orders via localAssignedOrderCH, localRequestCH and completedOrderCH.
// Also takes input from elevio on drv channels. Interacts with external timer on doorTimerStartCH and doorTimerFinishedCH
func Fsm(id string, localAssignedOrderCH, buttonPressCH, completedOrderCH chan elevio.ButtonEvent, fsmToOrdersCH chan elevator.Elevator) {

	drvButtonsCh := make(chan elevio.ButtonEvent)
	drvFloorsCh := make(chan int)
	drvObstrCh := make(chan bool)
	drvStopCh := make(chan bool)

	doorTimerStartCh := make(chan float64)
	doorTimerFinishedCh := make(chan bool)

	floorArrivalCh := make(chan bool)
	motorTimoutStartCh := make(chan bool)
	motorStopCh := make(chan bool)

	go elevio.PollButtons(drvButtonsCh)
	go elevio.PollFloorSensor(drvFloorsCh)
	go elevio.PollObstructionSwitch(drvObstrCh)
	go elevio.PollStopButton(drvStopCh) //kanskje implementere stop?

	//denne bør ha annet navn
	go timer.Timer(doorTimerStartCh, doorTimerFinishedCh)

	go timer.MotorStopTimer(floorArrivalCh, motorTimoutStartCh, motorStopCh)

	elev := elevator.Elevator_init(N_floors, N_buttons, id)

	for floor := range N_floors {
		for btn := range N_buttons {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
			elevio.SetDoorOpenLamp(false)
		}
	}


	floor := elevio.GetFloor()
	if floor == -1 {
		initBetweenFloors(&elev)
		current_floor := <-drvFloorsCh
		floorArrival(&elev, current_floor, doorTimerStartCh, floorArrivalCh, motorTimoutStartCh, completedOrderCH)
	} else {
		elev.Floor = floor
		elevio.SetFloorIndicator(floor)
	}

	// fmt.Println("startFloor: ",elev.Floor)

	//trenger kanskje ikke denne
	fsmToOrdersCH <- deepcopy.DeepCopyElevatorStruct(elev)

	for {
		fsmToOrdersCH <- deepcopy.DeepCopyElevatorStruct(elev)
		select {
		//Inputs (buttons pressed) on each elevator is channeled to their respective local request
		case button_input := <-drvButtonsCh:
			// fmt.Println("btn")
			buttonPressCH <- button_input

		//When an assigned order on a local elevator is channeled, it is set as an order to requestButtonPress that makes the elevators move
		case Order := <-localAssignedOrderCH:
			// fmt.Println("ordr press")
			requestButtonPress(&elev, Order.Floor, Order.Button, doorTimerStartCh, floorArrivalCh, motorTimoutStartCh, completedOrderCH)

		case current_floor := <-drvFloorsCh:
			floorArrival(&elev, current_floor, doorTimerStartCh, floorArrivalCh, motorTimoutStartCh, completedOrderCH)
			// fmt.Printf("drvFloorsCh: %v", current_floor)
		case obstruction := <-drvObstrCh:
			// fmt.Println("obstr")
			if obstruction {
				//clear local orders

				elev.Obstructed = true
				elev = clearLocalHallOrders(elev)
				fmt.Println("Obstruction switch activated")
				// (*elev).Behaviour = elevator.EB_Unavailable
			} else {
				elev.Obstructed = false
				// (*elev).Behaviour = elevator.EB_DoorOpen
				doorTimerStartCh <- elev.Config.DoorOpenDuration_s
			}
		case <-motorStopCh:
			fmt.Println("motorstop")
			elev.MotorStop = true
			elev = clearLocalHallOrders(elev)

		case <-doorTimerFinishedCh:
			if !elev.Obstructed && (elev.Behaviour != elevator.EB_Moving) {
				DoorTimeout(&elev, doorTimerStartCh, floorArrivalCh, motorTimoutStartCh, completedOrderCH)
				// fmt.Println("drv_doortimer timed out")
				// DoorTimeout(&elev, doorTimerStartCh, floorArrivalCh, motorTimoutStartCh, completedOrderCH)
			}
			// default:
			//non blocking
		}
	}
}
