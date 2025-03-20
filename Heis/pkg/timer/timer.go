package timer

import (
	// "fmt"
	"fmt"
	"time"
)

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
	timeoutInterval := 7 * time.Second
	// fmt.Println("start")
	// select {
	// case <- time.After(timeoutInterval):
	// 	motorStopCh <- true
	// case <- floorArrivalCh:
	// }
	// fmt.Println("finnish")

	fmt.Println("Timer started")


	timer := time.NewTimer(timeoutInterval)
	// if !timer.Stop() {
	// 	<-timer.C
	// }
	for {
		select {
		case <-motorTimeoutStartCh:
			// fmt.Println("bef rest")
			timer.Reset(timeoutInterval)
			// fmt.Println("adt rest")

		case <-floorArrivalCh:
			if !timer.Stop() {
				<-timer.C
			// break active
			}

		case <-timer.C:
			motorStopTimeoutCh <- true
		}
	}

}
