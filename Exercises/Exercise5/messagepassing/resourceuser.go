package main

import (
	"time"
)

func resourceUser(cfg ResourceUserConfig, take chan Resource, giveBack chan Resource) {

	time.Sleep(time.Duration(cfg.release) * tick)

	executionStates[cfg.id] = waiting
	res := <-take

	executionStates[cfg.id] = executing

	time.Sleep(time.Duration(cfg.execution) * tick)
	res.value = append(res.value, cfg.id)
	giveBack <- res

	executionStates[cfg.id] = done
}
