package timer

import (
	// "fmt"
	"fmt"
	"time"
)

//Bør skrive om dette til doortimer
// Takes in timer duration on timer start and sends on timer End when timer is finished. If timer start is sent twice before done, the timer is reset
func Timer(TimerStart chan float64, TimerEnd chan bool) {

	for {
		select {
		case t := <-TimerStart:
			timer := time.NewTimer(time.Second * time.Duration(t))
			for {
				select {
				case t := <-TimerStart:
					timer.Reset(time.Second * time.Duration(t))

				case <-timer.C:
					TimerEnd <- true
					break
				}
			}
		}
	}
}

func MotorStopTimer(floorArrivalCh, motorTimeoutStartCh, motorStopTimeoutCh chan bool) {
	//denne bør kanskje tas inn via config eller liknende
	//mindre enn 4 sekunder funker ikke på sim
	// timeoutInterval := 4 * time.Second
	timeoutInterval := 3900 * time.Millisecond

	fmt.Println("Timer started")


	timer := time.NewTimer(timeoutInterval)
	for {
		select {
		case <-motorTimeoutStartCh:
			timer.Reset(timeoutInterval)

		case <-floorArrivalCh:
			if !timer.Stop() {
				<-timer.C
			}

		case <-timer.C:
			motorStopTimeoutCh <- true
		}
	}

}
