package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) HandleCallback(update tg.Update) error {
	data := update.CallbackData()

	// Разбираем данные кнопки ("approve:123456789")
	firstPart, secondPart, err := bot.parseCallbackData(data)
	if err != nil {
		return fmt.Errorf("Can not parse callback data: %s", err)
	}

	if firstPart == "" {
		firstPart = data
	}

	switch firstPart {
	case approve:
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
			update.CallbackQuery.Message.Chat.ID,
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
	case reject:
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
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"❌ Пользователь отклонён.",
			nil,
		)
		if err != nil {
			return fmt.Errorf("Can not edit msg: %s", err)
		}
	case toCameras:
		cameras, err := bot.camerasAPI.GetAllCameras()
		if err != nil {
			return fmt.Errorf("Can not get all cameras: %s", err)
		}

		camerasMarkup := bot.getCameraButtons(cameras)
		_, err = bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			listCameras,
			&camerasMarkup,
		)
		if err != nil {
			return fmt.Errorf("Can not send msg: %s", err)
		}
	case toUsers:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, toMain),
			),
		)

		_, err = bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			addCameraAuto,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toAdd:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, toMain),
			),
		)

		_, err = bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			addCameraMac,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toMain:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(listCameras, toCameras),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(addCameraAuto, toUsers),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(addCameraMac, toAdd),
			),
		)
		_, err := bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			menu,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toCamera:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Сделать фото", fmt.Sprintf("%s:%s", toMakePhoto, secondPart)),
				tg.NewInlineKeyboardButtonData("Сделать видео", fmt.Sprintf("%s:%s", toChooseDuration, secondPart)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(infoCamera, fmt.Sprintf("%s:%s", toCameraInfo, secondPart)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, toCameras),
			),
		)
		_, err := bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			menu,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toMakePhoto:
		msgId := update.CallbackQuery.Message.MessageID

		if err := bot.Processing(update, secondPart, 0, msgId); err != nil {
			return fmt.Errorf("Can not process: %s", err)
		}
	case toChooseDuration:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("5 секунд", fmt.Sprintf("%s:%s/%d", toMakeVideo, secondPart, 5)),
				tg.NewInlineKeyboardButtonData("10 секунд", fmt.Sprintf("%s:%s/%d", toMakeVideo, secondPart, 10)),
				tg.NewInlineKeyboardButtonData("15 секунд", fmt.Sprintf("%s:%s/%d", toMakeVideo, secondPart, 15)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("30 секунд", fmt.Sprintf("%s:%s/%d", toMakeVideo, secondPart, 30)),
				tg.NewInlineKeyboardButtonData("45 секунд", fmt.Sprintf("%s:%s/%d", toMakeVideo, secondPart, 45)),
				tg.NewInlineKeyboardButtonData("60 секунд", fmt.Sprintf("%s:%s/%d", toMakeVideo, secondPart, 60)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, fmt.Sprintf("%s:%s", toCamera, secondPart)),
			),
		)

		_, err := bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"Выбор продолжительности записи видео:",
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toMakeVideo:
		msgId := update.CallbackQuery.Message.MessageID
		parts := strings.Split(secondPart, "/")

		duration, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("Can not parse duration: %s", err)
		}

		if err := bot.Processing(update, parts[0], duration, msgId); err != nil {
			return fmt.Errorf("Can not process: %s", err)
		}
	case toCameraInfo:
		msgId := update.CallbackQuery.Message.MessageID
		text := "Информация о камере\n" +
			"Название камеры: %s\n" +
			"RTSP: %s\n" +
			"IP:" +
			"Мак-адрес:"

		cameraInfo, err := bot.getCameraInfo(secondPart)
		if err != nil {
			return fmt.Errorf("Can not get camera info: %s", err)
		}

		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, fmt.Sprintf("%s:%s", toCamera, secondPart)),
			),
		)

		_, err = bot.EditMessage(
			update.CallbackQuery.Message.Chat.ID,
			msgId,
			fmt.Sprintf(text, cameraInfo.Name, cameraInfo.Rtsp),
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not edit msg: %s", err)
		}
	}

	return nil
}

// Разбитие строки "approve:123456789" на 2 части
func (bot *Bot) parseCallbackData(data string) (string, string, error) {
	parts := strings.Split(data, ":")

	var firstPart string
	var secondPart string
	if len(parts) == 1 {
		firstPart = parts[0]
	} else if len(parts) > 1 {
		firstPart = parts[0]
		secondPart = parts[1]
	} else {
		return "", "", ErrInvalidCallbackData
	}

	return firstPart, secondPart, nil
}

// Кнопка для создания того количества камер, которые есть в БД
func (bot *Bot) getCameraButtons(cameras []models.Camera) tg.InlineKeyboardMarkup {
	var buttons [][]tg.InlineKeyboardButton

	for _, camera := range cameras {
		button := tg.NewInlineKeyboardButtonData(camera.Name, fmt.Sprintf("%s:%s", toCamera, camera.Name))
		buttons = append(buttons, tg.NewInlineKeyboardRow(button))
	}
	back := tg.NewInlineKeyboardButtonData(back, toMain)
	buttons = append(buttons, tg.NewInlineKeyboardRow(back))

	return tg.NewInlineKeyboardMarkup(buttons...)
}

// Создает фото/видео
func (bot *Bot) makeFile(record models.Record) (int, error) {
	var tmp string
	var text string

	if record.Duration == nil {
		tmp = "capture"
		text = "Фото записывается. Оно будет готово через: 5 секунд."
	} else {
		tmp = "record"
		text = fmt.Sprintf("Видео записывается. Оно будет готово через: %d секунд.", *record.Duration+5)
	}

	// делает фото/видео
	if err := bot.camerasAPI.Capture(tmp, record); err != nil {
		return 0, fmt.Errorf("Can not make file: %s", err)
	}

	newMsg := tg.NewMessage(record.ChatID, text)
	msg, err := bot.tgAPI.Send(newMsg)
	if err != nil {
		return 0, fmt.Errorf("Can not send file: %s", err)
	}

	return msg.MessageID, nil
}

func (bot *Bot) Processing(update tg.Update, nameCamera string, duration int, msgId int) error {
	var record models.Record
	if duration == 0 {
		record = models.Record{
			ChatID:     update.CallbackQuery.From.ID,
			NameCamera: nameCamera,
			Duration:   nil,
		}
	} else {
		record = models.Record{
			ChatID:     update.CallbackQuery.From.ID,
			NameCamera: nameCamera,
			Duration:   &duration,
		}
	}

	if _, err := bot.makeFile(record); err != nil {
		return fmt.Errorf("Can not make file: %s", err)
	}

	del := tg.NewDeleteMessage(record.ChatID, msgId)
	if _, err := bot.tgAPI.Request(del); err != nil {
		return fmt.Errorf("Can not delete msg: %s", err)
	}

	msg := tg.NewMessage(record.ChatID, menu)
	buttons := tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(listCameras, toCameras),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(addCameraAuto, toUsers),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(addCameraMac, toAdd),
		),
	)

	msg.ReplyMarkup = buttons
	if _, err := bot.tgAPI.Send(msg); err != nil {
		return fmt.Errorf("Can not send msg: %s", err)
	}

	return nil
}

func (bot *Bot) getCameraInfo(nameCamera string) (models.Camera, error) {
	camera, err := bot.camerasAPI.GetCameraByName(nameCamera)
	if err != nil {
		return models.Camera{}, fmt.Errorf("Can not get camera: %s", err)
	}

	cameraInfo := models.Camera{
		Name: camera.Name,
		Rtsp: camera.Rtsp,
	}

	return cameraInfo, nil
}
