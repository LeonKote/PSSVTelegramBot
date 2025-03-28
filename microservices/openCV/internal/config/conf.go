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
}

const (
	useSSL = false

	address = "ADDRESS"

	endpoint        = "ENDPOINT"
	accessKeyID     = "ACCESS_KEY_ID"
	secretAccessKey = "SECRET_ACCESS_KEY"
	bucketName      = "BUCKET_NAME"
	rtsp            = "RTSP"
	filesApi        = "FILES_API"

	streamUrl = "STREAM_URL"
)

func MakeConfig() Config {
	return Config{
		Address: os.Getenv(address),

		Endpoint:        os.Getenv(endpoint),
		AccessKeyID:     os.Getenv(accessKeyID),
		SecretAccessKey: os.Getenv(secretAccessKey),
		UseSSL:          useSSL,
		BucketName:      os.Getenv(bucketName),
		FilesHost:       os.Getenv(filesApi),

		StreamUrl: os.Getenv(streamUrl),
	}
}
