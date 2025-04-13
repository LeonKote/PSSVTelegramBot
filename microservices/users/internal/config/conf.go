package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type Config struct {
	Logger zerolog.Logger

	Address   string
	BasicAuth string

	PgHost string
	PgPort uint64
	PgUser string
	PgDB   string
	PgPass string
}

const (
	address = "ADDRESS"
	auth    = "APIS_AUTH_BASIC"

	pgHost     = "POSTGRES_HOSTNAME"
	pgPort     = "POSTGRES_PORT"
	pgDB       = "POSTGRES_DATABASE"
	pgUser     = "POSTGRES_USER"
	pgPassword = "POSTGRES_PASSWORD"

	base = 10
	size = 64
)

func MakeConfig(logger zerolog.Logger) Config {
	port, err := strconv.ParseUint(os.Getenv(pgPort), base, size)
	if err != nil {
		logger.Panic().Msgf("Invalid postgres port: %s", err)
	}

	return Config{
		Logger: logger,

		Address:   os.Getenv(address),
		BasicAuth: os.Getenv(auth),

		PgHost: os.Getenv(pgHost),
		PgPort: port,
		PgUser: os.Getenv(pgUser),
		PgDB:   os.Getenv(pgDB),
		PgPass: os.Getenv(pgPassword),
	}
}
