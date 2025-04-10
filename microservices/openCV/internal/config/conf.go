package config

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	Logger zerolog.Logger

	Address string

	BasicAuth     string
	AuthForFfmpeg string

	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string

	NotificationsApi string
	CamerasApi       string
	StreamUrl        string
}

const (
	UseSSL = true

	address = "ADDRESS"

	basicAuth = "APIS_AUTH_BASIC"

	endpoint        = "ENDPOINT"
	accessKeyID     = "ACCESS_KEY_ID"
	secretAccessKey = "SECRET_ACCESS_KEY"
	bucketName      = "BUCKET_NAME"

	notificationsApi = "NOTIFICATIONS_API"
	camerasApi       = "CAMERAS_API"
	streamUrl        = "STREAM_URL"
)

func MakeConfig(logger zerolog.Logger) Config {
	auth := fmt.Sprintf("Authorization: Basic %s\r\n", base64.StdEncoding.EncodeToString([]byte(os.Getenv(basicAuth))))

	return Config{
		Logger: logger,

		Address:       os.Getenv(address),
		BasicAuth:     os.Getenv(basicAuth),
		AuthForFfmpeg: auth,

		Endpoint:        os.Getenv(endpoint),
		AccessKeyID:     os.Getenv(accessKeyID),
		SecretAccessKey: os.Getenv(secretAccessKey),
		UseSSL:          UseSSL,
		BucketName:      os.Getenv(bucketName),

		NotificationsApi: os.Getenv(notificationsApi),
		CamerasApi:       os.Getenv(camerasApi),
		StreamUrl:        os.Getenv(streamUrl),
	}
}
