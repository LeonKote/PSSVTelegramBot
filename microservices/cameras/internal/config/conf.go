package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/Impisigmatus/service_core/log"
)

type Config struct {
	Address    string
	BasicLogin string
	BasicPass  string

	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	FilesHost       string

	StreamUrl string

	PgHost string
	PgPort uint64
	PgUser string
	PgDB   string
	PgPass string
}

const (
	useSSL = false

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
	filesApi        = "FILES_API"

	streamUrl = "STREAM_URL"

	base = 10
	size = 64
)

func MakeConfig() Config {
	port, err := strconv.ParseUint(os.Getenv(pgPort), base, size)
	if err != nil {
		log.Panicf("Invalid postgres port: %s", err)
	}

	auth := strings.Split(os.Getenv(auth), ":")

	return Config{
		Address:    os.Getenv(address),
		BasicLogin: auth[0],
		BasicPass:  auth[1],

		Endpoint:        os.Getenv(endpoint),
		AccessKeyID:     os.Getenv(accessKeyID),
		SecretAccessKey: os.Getenv(secretAccessKey),
		UseSSL:          useSSL,
		BucketName:      os.Getenv(bucketName),
		FilesHost:       os.Getenv(filesApi),

		StreamUrl: os.Getenv(streamUrl),

		PgHost: os.Getenv(pgHost),
		PgPort: port,
		PgUser: os.Getenv(pgUser),
		PgDB:   os.Getenv(pgDB),
		PgPass: os.Getenv(pgPassword),
	}
}
