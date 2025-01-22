package fsm

import "Heis/pkg/elevio"

type ElevatorBehaviour int
const (
	EB_Idle ElevatorBehaviour
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int
const (
	CV_ALL ClearRequestVariant
	CV_InDirn
)

type Elevator struct {
	Floor int
	dirn elevio.MotorDirection
	requests [][]int
	behaviour ElevatorBehaviour

	config struct {
		clearRequestVariant ClearRequestVariant
		doorOpenDuration_s double
	}
}

func Fsm(drv_buttons chan elevio.ButtonEvent, drv_floors chan int, drv_obstr, drv_stop chan bool) {
// init state machine between floors
	var elevator Elevator

	fsm_init(&elevator)
	
	if(<- drv_floors == -1){
		initBetweenFloors()
	}

	for {
		select{
		case a := <- drv_buttons:
			floor func
		case a := <- drv_fldrv_floors:
		}
	}
}