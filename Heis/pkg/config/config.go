package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config defines the structure of the configuration
type Config struct {
	NumFloors  int `json:"NumFloors"`
	NumButtons int `json:"NumButtons"`
}

// Use to load Heis/config/elevator_params.json
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("could not decode config JSON: %v", err)
	}

	return config, nil
}

const NumFloors = 4
const NumButtons = 3
const DoorOpenDuration = 3
const StateUpdatePeriodMs = 500
const ElevatorStuckToleranceSec = 5
const ReconnectTimerSec = 3
const LocalElevator = 0


// type Direction int

// const (
// 	Up   Direction = 1
// 	Down Direction = -1
// 	Stop Direction = 0
// )

// type RequestState int

// const (
// 	None      RequestState = 0
// 	Order     RequestState = 1
// 	Comfirmed RequestState = 2
// 	Complete  RequestState = 3
// )

// type Behaviour int

// const (
// 	Idle        Behaviour = 0
// 	DoorOpen    Behaviour = 1
// 	Moving      Behaviour = 2
// 	Unavailable Behaviour = 3
// )

// type ButtonType int

// const (
// 	HallUp   ButtonType = 0
// 	HallDown ButtonType = 1
// 	Cab      ButtonType = 2
// )

// type Request struct {
// 	Floor  int
// 	Button ButtonType
// }

// type DistributorElevator struct {
// 	ID       string
// 	Floor    int
// 	Dir      Direction
// 	Requests [][]RequestState
// 	Behave   Behaviour
// }

// type CostRequest struct {
// 	Id         string
// 	Cost       int
// 	AssignedId string
// 	Req        Request
// }
