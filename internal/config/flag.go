package config

import "flag"

type Flags struct {
	RunAddress           string
	AccrualSystemAddress string
	DatabaseURI          string
}

func NewFlags() *Flags {
	runAddress := flag.String(
		"a",
		"",
		"адрес и порт запуска сервиса",
	)

	accrualSystemAddress := flag.String(
		"r",
		"",
		"адрес системы расчёта начислений",
	)

	databaseURI := flag.String(
		"d",
		"",
		"адрес подключения к базе данных",
	)

	flag.Parse()

	return &Flags{
		RunAddress:           *runAddress,
		AccrualSystemAddress: *accrualSystemAddress,
		DatabaseURI:          *databaseURI,
	}
}
