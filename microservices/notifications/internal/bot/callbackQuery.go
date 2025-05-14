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
	chatId := update.CallbackQuery.Message.Chat.ID
	msgId := update.CallbackQuery.Message.MessageID

	// Разбираем данные кнопки ("approve:123456789")
	firstPart, secondPart, err := bot.parseCallbackData(data)
	if err != nil {
		return fmt.Errorf("Can not parse callback data: %s", err)
	}

	if firstPart == "" {
		firstPart = data
	}

	switch firstPart {
	case approve: // разрешение на использование бота
		if err := bot.Approve(chatId, secondPart, &update); err != nil {
			return fmt.Errorf("Can not approve: %s", err)
		}
	case reject: // отказ в использовании бота
		if err := bot.Reject(chatId, secondPart, &update); err != nil {
			return fmt.Errorf("Can not reject: %s", err)
		}
	case toCameras: // кнопка перехода к выбору камеры
		cameras, err := bot.camerasAPI.GetAllCameras()
		if err != nil {
			return fmt.Errorf("Can not get all cameras: %s", err)
		}

		camerasMarkup := bot.getCameraButtons(cameras)
		_, err = bot.EditMessage(
			chatId,
			update.CallbackQuery.Message.MessageID,
			listCameras,
			&camerasMarkup,
		)
		if err != nil {
			return fmt.Errorf("Can not send msg: %s", err)
		}
	case toUsers: // кнопка перехода к просмотру пользователей
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, toMain),
			),
		)

		_, err = bot.EditMessage(
			chatId,
			update.CallbackQuery.Message.MessageID,
			addCameraAuto,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toAdd: // кнопка добавления камеры вручную
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, toMain),
			),
		)

		_, err = bot.EditMessage(
			chatId,
			update.CallbackQuery.Message.MessageID,
			"Введите RTSP-адрес камеры:",
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}

		UserStates[chatId] = "waitingForRTSP"
	case toAddCameraAuto: // кнопка добавления камеры автоматически
		if err := bot.autoAddCamera(chatId, &update); err != nil {
			return fmt.Errorf("Can not auto add camera: %s", err)
		}

	case toMain: // кнопка перехода в главное меню
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
		_, err := bot.EditMessage(
			chatId,
			update.CallbackQuery.Message.MessageID,
			menu,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toCamera: // кнопка перехода к камере
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Сделать фото", fmt.Sprintf("%s:%s", toMakePhoto, secondPart)),
				tg.NewInlineKeyboardButtonData("Сделать видео", fmt.Sprintf("%s:%s", toChooseDuration, secondPart)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(infoCamera, fmt.Sprintf("%s:%s", toCameraInfo, secondPart)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Изменить название камеры", fmt.Sprintf("%s:%s", toChangeNameOfCamera, secondPart)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Удалить камеру", fmt.Sprintf("%s:%s", toDeleteCamera, secondPart)),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, toCameras),
			),
		)
		_, err := bot.EditMessage(
			chatId,
			update.CallbackQuery.Message.MessageID,
			menu,
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toChangeNameOfCamera:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(back, fmt.Sprintf("%s:%s", toCamera, secondPart)),
			),
		)

		_, err := bot.EditMessage(
			chatId,
			update.CallbackQuery.Message.MessageID,
			"Введите новое название камеры:",
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}

		UserStates[chatId] = "waitingForNameOfCamera"
		CameraName[chatId] = secondPart
	case toDeleteCamera:
		if err := bot.camerasAPI.RemoveCamera(secondPart); err != nil {
			return fmt.Errorf("Can not remove camera: %s", err)
		}

		bot.EditMessage(chatId, msgId, "Камера удалена", nil)

		if err := bot.MakeNewMain(chatId); err != nil {
			return fmt.Errorf("Can not make new main: %s", err)
		}
	case toMakePhoto: // кнопка сделать фото
		if err := bot.Processing(update, secondPart, 0, msgId); err != nil {
			return fmt.Errorf("Can not process: %s", err)
		}
	case toChooseDuration: // кнопка выбора продолжительности записи видео
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
			chatId,
			update.CallbackQuery.Message.MessageID,
			"Выбор продолжительности записи видео:",
			&buttons,
		)
		if err != nil {
			return fmt.Errorf("Can not make button: %s", err)
		}
	case toMakeVideo: // кнопка сделать видео
		msgId := update.CallbackQuery.Message.MessageID
		parts := strings.Split(secondPart, "/")

		duration, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("Can not parse duration: %s", err)
		}

		if err := bot.Processing(update, parts[0], duration, msgId); err != nil {
			return fmt.Errorf("Can not process: %s", err)
		}
	case toCameraInfo: // кнопка информации о камере
		msgId := update.CallbackQuery.Message.MessageID
		text := "Информация о камере\n" +
			"Название камеры: %s\n" +
			"RTSP: %s\n"

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
			chatId,
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

// Запись фото/видео и загрузка в s3
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
			tg.NewInlineKeyboardButtonData(addCameraAuto, toAddCameraAuto),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(addCameraRtsp, toAdd),
		),
	)

	msg.ReplyMarkup = buttons
	if _, err := bot.tgAPI.Send(msg); err != nil {
		return fmt.Errorf("Can not send msg: %s", err)
	}

	return nil
}

// Получение информации о камере
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
