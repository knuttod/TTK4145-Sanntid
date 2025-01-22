package timer

import "time"

func Timer(drv_doorTimer chan float64) {
	for {
		select {
		case a := <-drv_doorTimer:
			time.Sleep(time.Second * time.Duration(a))
			drv_doorTimer <- 0.0
		}
	}
}
