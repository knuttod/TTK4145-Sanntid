package fsm

import "Heis/pkg/elevio"

// jonas
const N_floors = 4
const N_buttons = 3

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_ALL ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [N_floors][N_buttons]int
	Behaviour ElevatorBehaviour

	Config struct { //type?
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}

func Fsm(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool, drv_doorTimer chan float64) {
	// init state machine between floors
	var elevator Elevator
	fsm_init(&elevator)

	if <-drv_floors == -1 {
		initBetweenFloors(&elevator)
	}

	for {
		select {
		case a := <-drv_buttons:
			requestButtonPress(&elevator, a.Floor, a.Button, drv_doorTimer)
		case a := <-drv_floors:
			floorArrival(&elevator, a, drv_doorTimer)
		case <-drv_doorTimer:
			DoorTimeout(&elevator, drv_doorTimer)
		}
	}
}
