package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	Environment *EnvironmentSettings
	Flags       *Flags
}

type EnvironmentSettings struct {
	Integration *IntegrationSettings
	Database    *PGSettings
	Server      *ServerSettings
	JWT         *JWTSettings
}

func NewSettings() *Settings {
	envSettings := NewEnvironmentSettings()
	flags := NewFlags()

	return &Settings{
		Environment: envSettings,
		Flags:       flags,
	}
}

func NewEnvironmentSettings() *EnvironmentSettings {
	var settings EnvironmentSettings
	if err := envconfig.Process("", &settings); err != nil {
		panic("Failed to load settings: " + err.Error())
	}

	return &settings
}

func (s *Settings) GetServerAddress() string {
	// Если указана переменная окружения, то используется она
	if serverAddr := strings.TrimSpace(s.Environment.Server.Address); serverAddr != "" {
		return serverAddr
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if serverAddr := strings.TrimSpace(s.Flags.RunAddress); serverAddr != "" {
		return serverAddr
	}

	// Если нет ни переменной окружения, ни флага, то используются значения по умолчанию
	return DefaultServerAddress
}

func (s *Settings) GetDatabaseURI() string {
	// Если указана переменная окружения, то используется она
	if dbURI := strings.TrimSpace(s.Environment.Database.URI); dbURI != "" {
		return dbURI
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if dbURI := strings.TrimSpace(s.Flags.DatabaseURI); dbURI != "" {
		return dbURI
	}

	// Если нет ни переменной окружения, ни флага, то используются значения по умолчанию
	return DefaultPostgresDSN
}

func (s *Settings) GetAccrualSystemAddress() string {
	// Если указана переменная окружения, то используется она
	if addr := strings.TrimSpace(s.Environment.Integration.AccrualSystemAddress); addr != "" {
		return addr
	}

	// Если нет переменной окружения, но есть аргумент командной строки(флаг), то используется он
	if addr := strings.TrimSpace(s.Flags.AccrualSystemAddress); addr != "" {
		return addr
	}

	// Если нет ни переменной окружения, ни флага, то используются значения по умолчанию
	return DefaultAccrualSystemAddress
}
