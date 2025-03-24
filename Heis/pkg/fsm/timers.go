package fsm

import (
	"time"
)

//Bør skrive om dette til doortimer
// Takes in timer duration on timer start and sends on timer End when timer is finished. If timer start is sent twice before done, the timer is reset
func doorTimer(doorOpenTime time.Duration, doorTimerStartCh chan bool, doorTimerEndCh chan bool) {
	//doorOpenTime is expected to be in seconds

	timeoutInterval := time.Duration(doorOpenTime) * time.Second

	for {
		select {
		case <-doorTimerStartCh:
			timer := time.NewTimer(timeoutInterval)
			for {
				select {
				case <-doorTimerStartCh:
					timer.Reset(timeoutInterval)

				case <-timer.C:
					doorTimerEndCh <- true
					break
				}
			}
		}
	}
}

func motorStopTimer(motorStopTimeoutTime time.Duration, arrivedOnFloorCh, departureFromFloorCh, motorStopCh chan bool) {

	//denne bør kanskje være lang
	timer := time.NewTimer(motorStopTimeoutTime)
	for {
		select {

		case <-departureFromFloorCh:
			timer.Reset(motorStopTimeoutTime)

		case <-arrivedOnFloorCh:
			if !timer.Stop() {
				<-timer.C
			}

		case <-timer.C:
			departureFromFloorCh <- true
		}
	}
}
