package config

import (
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	Logger zerolog.Logger

	Address   string
	BasicAuth string

	CamerasUrl string
}

const (
	address    = "ADDRESS"
	auth       = "APIS_AUTH_BASIC"
	camerasUrl = "CAMERAS_API"
)

func MakeConfig(logger zerolog.Logger) Config {
	return Config{
		Logger:     logger,
		Address:    os.Getenv(address),
		BasicAuth:  os.Getenv(auth),
		CamerasUrl: os.Getenv(camerasUrl),
	}
}
