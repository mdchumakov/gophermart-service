package config

const DefaultAccrualSystemAddress = "localhost:8081"

type IntegrationSettings struct {
	AccrualSystemAddress string `envconfig:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualSystemTimeout int    `envconfig:"ACCRUAL_SYSTEM_TIMEOUT" default:"5"`
}
