package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/bot"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
			file := frame.File[len(path.Dir(os.Args[0]))+1:]
			line := frame.Line
			return "", fmt.Sprintf("%s:%d", file, line)
		},
	})
}

func main() {
	const (
		base        = 10
		size        = 64
		token       = "TOKEN"
		usersApi    = "USERS_API"
		camerasApi  = "CAMERAS_API"
		adminIdStr  = "ADMIN_CHAT_ID"
		addrUsers   = "ADDRESS_USERS_API"
		addrCameras = "ADDRESS_CAMERAS_API"
	)

	adminID, err := strconv.ParseInt(os.Getenv(adminIdStr), base, size)
	if err != nil {
		logrus.Fatalf("Ошибка конвертации ADMIN_CHAT_ID в int64: %v", err)
	}

	// Запуск бота
	bot, err := bot.NewBot(os.Getenv(token),
		adminID,
		os.Getenv(usersApi),
		os.Getenv(camerasApi),
		os.Getenv(addrUsers),
		os.Getenv(addrCameras))
	if err != nil {
		logrus.Errorf("Can not authorized bot: %s", err)
	}

	bot.Run()
}
