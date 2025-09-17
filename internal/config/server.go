package config

const DefaultServerAddress = "localhost:8080"

type ServerSettings struct {
	Address string `envconfig:"RUN_ADDRESS"`
}
