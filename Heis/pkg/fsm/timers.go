package fsm

import (
	"time"
)

// Timer for the door. Should run as a go-routine. Opening time of door is given as an input.
// Timer is started by sending on doorTimerStartCh and timeout of timer is sent on doorTimerEndCh
func doorTimer(doorOpenTime time.Duration, doorTimerStartCh chan bool, doorTimerEndCh chan bool) {

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

// Timer to detect motorstop. Should run as a go-routine. The timeout for detecting motorstop is given as an input. 
// Timer is started by sending on departureFromFloorCh (every time the elevator departures from floor).
// Timer is aborted if something is sent on arrivedOnFloorCh (every time an elevator reaches a floor).
// On a timeout, aka. a motorstop happend, it is sent on motorStopCh.
func motorStopTimer(motorStopTimeoutTime time.Duration, arrivedOnFloorCh, departureFromFloorCh, motorStopCh chan bool) {

	<- departureFromFloorCh
	timer := time.NewTimer(motorStopTimeoutTime)
	for {
		select {

		case <-departureFromFloorCh:
			timer.Reset(motorStopTimeoutTime)

		case <-arrivedOnFloorCh:
			//Tries to stop timer, if not possible take return value from timer.C
			if !timer.Stop() {
				<-timer.C
			}

		case <-timer.C:
			motorStopCh <- true
		}
	}
}
