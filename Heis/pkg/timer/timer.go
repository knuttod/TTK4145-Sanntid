package timer

import "time"

func Timer(Timer chan float64) {
	for {
		select {
		case t := <-Timer:
			time.Sleep(time.Second * time.Duration(t))
			Timer <- 0.0
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
