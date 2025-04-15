package env

import (
	"fmt"
	"github.com/caarlos0/env"
)

type ReminderThresholdConfig struct {
	ReminderThresholdForFreeTrial int `env:"REMINDER_THRESHOLD_FOR_FREE_TRIAL" envDefault:"3" description:"this is used to set ( to be remind in time) for free trial"`
	ReminderThresholdForLicense   int `env:"REMINDER_THRESHOLD_FOR_LICENSE" envDefault:"15" description:"this is used to set ( to be remind in time) for license"`
}

func GetThresholdReminderConfig() (*ReminderThresholdConfig, error) {
	cfg := &ReminderThresholdConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse thresholdRemainder for licenses: " + err.Error())
		return nil, err
	}
	return cfg, nil
}
