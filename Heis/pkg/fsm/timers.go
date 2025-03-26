package fsm

import (
	"time"
)

// Takes in timer duration on timer start and sends on timer End when timer is finished. If timer start is sent twice before done, the timer is reset
func doorTimer(doorOpenTime time.Duration, doorTimerStartCh chan bool, doorTimerEndCh chan bool) {
	//doorOpenTime is expected to be in seconds

	//tror dette er riktig/lurt
	<- doorTimerStartCh
	timer := time.NewTimer(doorOpenTime)
	for {
		select {
		case <-doorTimerStartCh:
			timer.Reset(doorOpenTime)
		case <-timer.C:
			doorTimerEndCh <- true
		}
	}
}

func motorStopTimer(motorStopTimeoutTime time.Duration, arrivedOnFloorCh, departureFromFloorCh, motorStopCh chan bool) {
	
	//tror dette er riktig/lurt
	<- departureFromFloorCh
	//denne bør kanskje være lang
	timer := time.NewTimer(motorStopTimeoutTime)
	for {
		select {

		case <-departureFromFloorCh:
			timer.Reset(motorStopTimeoutTime)

		case <-arrivedOnFloorCh:
			//Tries to stop timer, if not possible take return value from timer
			if !timer.Stop() {
				<-timer.C
			}

		case <-timer.C:
			motorStopCh <- true
		}
	}
}
