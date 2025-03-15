package bot

import (
	"log"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/api"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	logrus "github.com/sirupsen/logrus"
)

type Bot struct {
	tgAPI      *tg.BotAPI
	usersAPI   api.UsersAPI
	camerasAPI api.CameraAPI
	admin      int64
}

const (
	toCameras = "redirect_to_cameras"
	toUsers   = "redirect_to_users"
	toMain    = "redirect_to_main"
	toCamera  = "redirect_to_camera"
	toAdd     = "redirect_to_add"
)

var LastActions = make(map[int64]int)

func NewBot(token string, admin int64, usersApi, camerasApi string, addrUsers, addrCameras string) (*Bot, error) {
	botApi, err := tg.NewBotAPI(token)
	if err != nil {
		return &Bot{}, err
	}

	usersAPI := api.NewUsersApi(addrUsers)
	camerasAPI := api.NewCameraApi(addrCameras)
	bot := Bot{
		tgAPI:      botApi,
		usersAPI:   *usersAPI,
		camerasAPI: *camerasAPI,
		admin:      admin,
	}

	logrus.Info("Бот успешно запущен.")

	return &bot, nil
}

func (bot *Bot) Run() {
	updater := tg.NewUpdate(0)
	updater.Timeout = 60
	updates := bot.tgAPI.GetUpdatesChan(updater)

	for update := range updates {
		if err := bot.handle(update); err != nil {
			logrus.Errorf("Invalid update: %s", err)
			continue
		}
	}
}

func (bot *Bot) handle(update tg.Update) error {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		bot.HandleMessage(update.Message)
	}

	if update.CallbackQuery != nil {
		bot.HandleCallback(update)
	}

	return nil
}
