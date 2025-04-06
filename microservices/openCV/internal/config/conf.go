package config

import (
	"os"
)

type Config struct {
	Address string

	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	FilesHost       string

	StreamUrl string

	NotificationsApi string
}

const (
	useSSL = true

	address = "ADDRESS"

	endpoint        = "ENDPOINT"
	accessKeyID     = "ACCESS_KEY_ID"
	secretAccessKey = "SECRET_ACCESS_KEY"
	bucketName      = "BUCKET_NAME"

	streamUrl = "STREAM_URL"

	notificationsApi = "NOTIFICATIONS_API"
)

func MakeConfig() Config {
	return Config{
		Address: os.Getenv(address),

		Endpoint:        os.Getenv(endpoint),
		AccessKeyID:     os.Getenv(accessKeyID),
		SecretAccessKey: os.Getenv(secretAccessKey),
		UseSSL:          useSSL,
		BucketName:      os.Getenv(bucketName),

		StreamUrl:        os.Getenv(streamUrl),
		NotificationsApi: os.Getenv(notificationsApi),
	}
}
