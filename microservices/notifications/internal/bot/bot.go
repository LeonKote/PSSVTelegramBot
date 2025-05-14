package bot

import (
	"fmt"
	"strconv"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	"github.com/Ullaakut/nmap"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Bot struct {
	tgAPI      *tg.BotAPI
	usersAPI   api.UsersAPI
	camerasAPI api.CameraAPI
	admin      int64
	log        zerolog.Logger
	cfg        config.Config
}

var (
	LastActions = make(map[int64]int)
	UserStates  = make(map[int64]string)
	UserRtsp    = make(map[int64]string)
	CameraName  = make(map[int64]string)
)

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
		cfg:        cfg,
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

func (bot *Bot) autoAddCamera(chatId int64, update *tg.Update) error {
	_, err := bot.EditMessage(
		chatId,
		update.CallbackQuery.Message.MessageID,
		"Идёт поиск камеры в сети. Пожалуйста, подождите.",
		nil,
	)
	if err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	}

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(bot.cfg.TargetAddr),
		nmap.WithPorts("554"),
	)
	if err != nil {
		return fmt.Errorf("Can not create scanner: %s", err)
	}

	// Запускаем сканирование
	result, wrn, err := scanner.Run()
	if err != nil || wrn != nil {
		return fmt.Errorf("Can not scan: %s", err)
	}

	// Печатаем результат сканирования
	text := "Найдены новые камеры"

	cameras, err := bot.camerasAPI.GetAllCameras()
	if err != nil {
		return fmt.Errorf("Can not get all cameras: %s", err)
	}

	var isCameraAdded bool
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		ok := false
		var rtsp string
		for _, port := range host.Ports {
			if port.Status() != nmap.Open {
				continue
			}

			for _, camera := range cameras {
				rtsp = fmt.Sprintf("rtsp://%s:%d/ucast/11", host.Addresses[0], port.ID)
				if camera.Rtsp == rtsp {
					ok = true
					break
				}

			}
		}

		if !ok {
			if rtsp == "" {
				continue
			}
			reqId := uuid.New().String()
			if err := bot.camerasAPI.AddCamera(models.Camera{
				Name: reqId,
				Rtsp: rtsp,
			}); err != nil {
				_, err := bot.EditMessage(
					chatId,
					update.CallbackQuery.Message.MessageID,
					"Что-то пошло не так :(\n Попробуйте заново.",
					nil,
				)
				if err != nil {
					return fmt.Errorf("Can not edit msg: %s", err)
				}

				return fmt.Errorf("Can not add camera: %s", err)
			}
			isCameraAdded = true
		}
	}

	if !isCameraAdded {
		text = "Ничего не найдено."
	}

	_, err = bot.EditMessage(
		chatId,
		update.CallbackQuery.Message.MessageID,
		text,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	}

	buttons := tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(listCameras, toCameras),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(addCameraAuto, toAddCameraAuto),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(addCameraRtsp, toAdd),
		),
	)

	if err := bot.SendMessage(
		chatId,
		menu,
		&buttons,
	); err != nil {
		return fmt.Errorf("Can not make button: %s", err)
	}
	return nil
}

func (bot *Bot) Approve(chatId int64, secondPart string, update *tg.Update) error {
	userID, err := strconv.ParseInt(secondPart, 10, 64)
	if err != nil {
		return fmt.Errorf("Can not parse user id: %s", err)
	}

	user, err := bot.usersAPI.GetUserByChatID(userID)
	if err != nil {
		return fmt.Errorf("Can not get user by chat_id: %s", err)
	}

	ok, err := bot.usersAPI.UpdateUser(models.User{
		Chat_ID:  userID,
		Username: user.Username,
		Name:     user.Name,
		Is_Admin: false,
		Status:   approved,
	})
	if err != nil || !ok {
		return fmt.Errorf("Can not update user: %s", err)
	}

	messageID, exists := LastActions[userID]
	if !exists {
		return fmt.Errorf("Can not find message by user_id: %s", err)
	}

	newMsg, err := bot.EditMessage(
		userID,
		messageID,
		"🎉 Вам одобрили доступ!",
		nil,
	)
	if err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	} else {
		LastActions[userID] = newMsg.MessageID
	}

	_, err = bot.EditMessage(
		chatId,
		update.CallbackQuery.Message.MessageID,
		"✅ Пользователь одобрен!",
		nil,
	)
	if err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	}

	buttons := tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(start, toMain),
		),
	)

	text := "Нажмите кнопку \"Начать\", чтобы начать испольование бота."
	if err = bot.SendMessage(userID, text, &buttons); err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	}

	return nil
}

func (bot *Bot) Reject(chatId int64, secondPart string, update *tg.Update) error {
	userID, err := strconv.ParseInt(secondPart, 10, 64)
	if err != nil {
		return fmt.Errorf("Can not parse user id: %s", err)
	}

	user, err := bot.usersAPI.GetUserByChatID(userID)
	if err != nil {
		return fmt.Errorf("Can not get user by chat_id: %s", err)
	}

	ok, err := bot.usersAPI.UpdateUser(models.User{
		Chat_ID:  userID,
		Username: user.Username,
		Name:     user.Name,
		Is_Admin: false,
		Status:   rejected,
	})
	if err != nil || !ok {
		return fmt.Errorf("Can not update user: %s", err)
	}

	// Сохранение ID последнего сообщения от пользователя
	messageID, exists := LastActions[userID]
	if !exists {
		return fmt.Errorf("Can not find message by user_id: %s", err)
	}

	newMsg, err := bot.EditMessage(
		userID,
		messageID,
		"🚫 Ваш запрос отклонён.",
		nil,
	)
	if err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	} else {
		LastActions[userID] = newMsg.MessageID
	}

	_, err = bot.EditMessage(
		chatId,
		update.CallbackQuery.Message.MessageID,
		"❌ Пользователь отклонён.",
		nil,
	)
	if err != nil {
		return fmt.Errorf("Can not edit msg: %s", err)
	}

	return nil
}
