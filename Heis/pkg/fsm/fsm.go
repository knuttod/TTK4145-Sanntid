package fsm

import (
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"fmt"
)

// define in config
const N_floors = 4
const N_buttons = 3

func Fsm(elev *elevator.Elevator, drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr,
	drv_stop chan bool, drv_doorTimerStart chan float64, drv_doorTimerFinished chan bool,
	id string, localAssignedOrder, localRequest chan elevio.ButtonEvent, completedOrderCH chan elevio.ButtonEvent) {

	if elevio.GetFloor() == -1 {
		initBetweenFloors(elev)
	}

	for {
		select {
		//Inputs (buttons pressed) on each elevator is channeled to their respective local request
		case button_input := <-drv_buttons:
			localRequest <- button_input

		//When an assigned order on a local elevator is channeled, it is set as an order to requestButtonPress that makes the elevators move
		case Order := <-localAssignedOrder:
			requestButtonPress(elev, Order.Floor, Order.Button, drv_doorTimerStart, completedOrderCH)

		case current_floor := <-drv_floors:
			floorArrival(elev, current_floor, drv_doorTimerStart, completedOrderCH)

			fmt.Printf("drv_floors: %v", current_floor)
		case obstruction := <-drv_obstr:
			if obstruction {
				(*elev).Obstructed = true
			} else {
				(*elev).Obstructed = false
				drv_doorTimerStart <- (*elev).Config.DoorOpenDuration_s
			}

		case <-drv_doorTimerFinished:
			if !elev.Obstructed {
				DoorTimeout(elev, drv_doorTimerStart, completedOrderCH)
				fmt.Println("drv_doortimer timed out")
				DoorTimeout(elev, drv_doorTimerStart, completedOrderCH)
			}

		}
	}
}
