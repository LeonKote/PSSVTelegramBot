package config

import (
	"os"
	"strings"

	"github.com/Impisigmatus/service_core/log"
)

type Config struct {
	Address    string
	BasicLogin string
	BasicPass  string

	CamerasUrl string
}

const (
	address    = "ADDRESS"
	auth       = "APIS_AUTH_BASIC"
	camerasUrl = "CAMERAS_API"
)

func MakeConfig() Config {
	auth := strings.Split(os.Getenv(auth), ":")

	log.Infof("Auth: %s:%s", auth[0], auth[1])

	return Config{
		Address:    os.Getenv(address),
		BasicLogin: auth[0],
		BasicPass:  auth[1],
		CamerasUrl: os.Getenv(camerasUrl),
	}
}
