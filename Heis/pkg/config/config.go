package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config defines the structure of the configuration
type Config struct {
	NumFloors        int           `json:"NumFloors"`
	NumBtns    		 int           `json:"NumBtns"`
	TravelTime       int           `json:"TravelTime"`
	DoorOpenDuration time.Duration `json:"DoorOpenDuration"`
	MotorStopTimeout time.Duration `json:"motorStopTimeout"`
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
