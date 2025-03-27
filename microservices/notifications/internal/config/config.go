package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AdminId int64

	Token string

	UsersApi   string
	CamerasApi string

	Address    string
	BasicLogin string
	BasicPass  string
}

const (
	adminIdStr = "ADMIN_CHAT_ID"

	token = "TOKEN"

	usersApi   = "USERS_API"
	camerasApi = "CAMERAS_API"

	address = "ADDRESS"
	auth    = "APIS_AUTH_BASIC"

	base = 10
	size = 64
)

func MakeConfig() *Config {
	adminID, err := strconv.ParseInt(os.Getenv(adminIdStr), base, size)
	if err != nil {
		log.Fatalf("Ошибка конвертации ADMIN_CHAT_ID в int64: %v", err)
	}

	auth := strings.Split(os.Getenv(auth), ":")

	return &Config{
		AdminId:    adminID,
		Token:      os.Getenv(token),
		UsersApi:   os.Getenv(usersApi),
		CamerasApi: os.Getenv(camerasApi),
		Address:    os.Getenv(address),
		BasicLogin: auth[0],
		BasicPass:  auth[1],
	}
}
