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

	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	NotifyHost      string

	PgHost string
	PgPort uint64
	PgUser string
	PgDB   string
	PgPass string
}

const (
	UseSSL = true

	address = "ADDRESS"
	auth    = "APIS_AUTH_BASIC"

	pgHost     = "POSTGRES_HOSTNAME"
	pgPort     = "POSTGRES_PORT"
	pgDB       = "POSTGRES_DATABASE"
	pgUser     = "POSTGRES_USER"
	pgPassword = "POSTGRES_PASSWORD"

	endpoint        = "ENDPOINT"
	accessKeyID     = "ACCESS_KEY_ID"
	secretAccessKey = "SECRET_ACCESS_KEY"
	bucketName      = "BUCKET_NAME"
	rtsp            = "RTSP"

	notifyApi = "NOTIFY_API"

	base = 10
	size = 64
)

func MakeConfig(log zerolog.Logger) Config {
	port, err := strconv.ParseUint(os.Getenv(pgPort), base, size)
	if err != nil {
		log.Panic().Msgf("Invalid postgres port: %s", err)
	}

	return Config{
		Logger: log,

		Address:   os.Getenv(address),
		BasicAuth: os.Getenv(auth),

		Endpoint:        os.Getenv(endpoint),
		AccessKeyID:     os.Getenv(accessKeyID),
		SecretAccessKey: os.Getenv(secretAccessKey),
		UseSSL:          UseSSL,
		BucketName:      os.Getenv(bucketName),
		NotifyHost:      os.Getenv(notifyApi),

		PgHost: os.Getenv(pgHost),
		PgPort: port,
		PgUser: os.Getenv(pgUser),
		PgDB:   os.Getenv(pgDB),
		PgPass: os.Getenv(pgPassword),
	}
}
