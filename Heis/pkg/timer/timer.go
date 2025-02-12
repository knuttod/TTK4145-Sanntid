package timer

import "time"

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

func TimerFinished(Timer chan float64) {
	var finished bool = false
	for finished == false {
		select {
		case t := <-Timer:
			if t == 0.0 {
				finished = true
			}
		}
	}
}
