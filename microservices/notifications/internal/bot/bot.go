package bot

import (
	"fmt"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/config"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type Bot struct {
	tgAPI      *tg.BotAPI
	usersAPI   api.UsersAPI
	camerasAPI api.CameraAPI
	admin      int64
	log        zerolog.Logger
}

var LastActions = make(map[int64]int)

func NewBot(log zerolog.Logger, cfg config.Config) *Bot {
	botApi, err := tg.NewBotAPI(cfg.Token)
	if err != nil {
		return &Bot{}
	}

	usersAPI := api.NewUsersApi(cfg)
	camerasAPI := api.NewCameraApi(cfg)
	bot := Bot{
		tgAPI:      botApi,
		usersAPI:   *usersAPI,
		camerasAPI: *camerasAPI,
		admin:      cfg.AdminId,
		log:        log,
	}

	log.Info().Msg("Бот успешно запущен.")

	return &bot
}

func (bot *Bot) Run() {
	updater := tg.NewUpdate(0)
	updater.Timeout = 60
	updates := bot.tgAPI.GetUpdatesChan(updater)

	for update := range updates {
		if err := bot.handle(update); err != nil {
			bot.log.Error().Msgf("Invalid update: %s", err)
			continue
		}
	}
}

func (bot *Bot) handle(update tg.Update) error {
	if update.Message != nil {
		bot.log.Info().Msgf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if err := bot.HandleMessage(update.Message); err != nil {
			return fmt.Errorf("Can not handle message: %s", err)
		}
	}

	if update.CallbackQuery != nil {
		if err := bot.HandleCallback(update); err != nil {
			return fmt.Errorf("Can not handle callback: %s", err)
		}
	}

	return nil
}

func (bot *Bot) EditMessage(chatId int64, messageId int, desc string, buttons *tg.InlineKeyboardMarkup) (tg.Message, error) {
	msg := tg.NewEditMessageText(chatId, messageId, desc)

	msg.ReplyMarkup = buttons

	newMsg, err := bot.tgAPI.Send(msg)
	if err != nil {
		return tg.Message{}, fmt.Errorf("Can not send msg: %s", err)
	}

	return newMsg, nil
}

func (bot *Bot) SendMessage(chatId int64, desc string, buttons *tg.InlineKeyboardMarkup) error {
	msg := tg.NewMessage(chatId, desc)

	msg.ReplyMarkup = buttons

	if _, err := bot.tgAPI.Send(msg); err != nil {
		return fmt.Errorf("Can not send msg: %s", err)
	}

	return nil
}
