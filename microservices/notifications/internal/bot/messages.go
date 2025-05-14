package bot

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) HandleMessage(msg *tg.Message) error {
	chatId := msg.Chat.ID
	if msg.Text == "/start" || msg.Text == start {
		if UserRtsp[chatId] != "" || UserStates[chatId] != "" || CameraName[chatId] != "" {
			delete(UserRtsp, chatId)
			delete(UserStates, chatId)
			delete(CameraName, chatId)
		}

		user, err := bot.usersAPI.GetUserByChatID(chatId)
		if err != nil {
			return fmt.Errorf("Invalid get user by chat id: %s", err)
		}

		if chatId == bot.admin && user == (models.User{}) {
			ok, err := bot.usersAPI.AddUser(models.User{
				Chat_ID:  chatId,
				Username: msg.From.UserName,
				Name:     msg.From.FirstName,
				Is_Admin: true,
				Status:   approved,
			})
			if err != nil || !ok {
				return fmt.Errorf("Invalid add user: %s", err)
			}
		}

		switch user.Status {
		case approved:
			buttons := tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData(start, toMain),
				),
			)

			if err := bot.SendMessage(msg.Chat.ID, "Добро пожаловать! Нажмите кнопку, чтобы начать испольование бота.", &buttons); err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}
		case pending:
			msg := tg.NewMessage(msg.Chat.ID, "Ожидает подтверждения.")
			if _, err := bot.tgAPI.Send(msg); err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}
		case rejected:
			msg := tg.NewMessage(msg.Chat.ID, "❌ Вам отказано в доступе.")
			if _, err := bot.tgAPI.Send(msg); err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}

		}

		if user == (models.User{}) {
			ok, err := bot.usersAPI.AddUser(models.User{
				Chat_ID:  chatId,
				Username: msg.From.UserName,
				Name:     msg.From.FirstName,
				Is_Admin: false,
				Status:   pending,
			})
			if err != nil || !ok {
				return fmt.Errorf("Invalid add user: %s", err)
			}

			msgToAdmin := tg.NewMessage(bot.admin,
				fmt.Sprintf("❗ Новый пользователь запросил доступ:\n\n👤 %s (@%s)\n🔢 ID: %d",
					fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName),
					msg.From.UserName,
					msg.From.ID))

			fromChatID := strconv.Itoa(int(chatId))
			keyboard := tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData("✅ Добавить", fmt.Sprintf("%s:%s", approve, fromChatID)),
					tg.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("%s:%s", reject, fromChatID)),
				),
			)

			msgToAdmin.ReplyMarkup = keyboard
			if _, err := bot.tgAPI.Send(msgToAdmin); err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}

			// Уведомляем пользователя
			newMsg, err := bot.tgAPI.Send(tg.NewMessage(chatId, "Ваш запрос отправлен администратору на проверку. Ожидайте."))
			if err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}
			LastActions[chatId] = newMsg.MessageID
		}
	} else if UserStates[chatId] == "waitingForRTSP" {
		rtspAddress := msg.Text
		if strings.Contains(rtspAddress, "rtsp://") {
			// Сохраняем RTSP-адрес в состояние пользователя
			UserStates[chatId] = "waitingForName" // Переходим ко следующему шагу
			UserRtsp[chatId] = rtspAddress

			// Запрашиваем ввод названия камеры
			if err := bot.SendMessage(chatId, "Теперь введите название камеры:", nil); err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}
		} else {
			// Если введен некорректный RTSP-адрес
			if err := bot.SendMessage(chatId, "Пожалуйста, введите правильный RTSP-адрес.", nil); err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}
		}
		return nil
	} else if UserStates[chatId] == "waitingForName" {
		cameraName := msg.Text

		err := bot.camerasAPI.AddCamera(models.Camera{
			Name: cameraName,
			Rtsp: UserRtsp[chatId],
		})
		if err != nil {
			delete(UserRtsp, chatId)
			delete(UserStates, chatId)

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

			err = bot.SendMessage(
				chatId,
				"Камера не была добавлена. Попробуйте позднее",
				&buttons,
			)
			if err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}

			return fmt.Errorf("Can not add camera: %s", err)
		}

		// Подтверждаем сохранение
		err = bot.SendMessage(
			chatId,
			fmt.Sprintf("Камера успешно добавлена!\nНазвание: %s\nRTSP-адрес: %s\n",
				cameraName,
				UserRtsp[chatId]),
			nil,
		)
		if err != nil {
			return fmt.Errorf("Can not send msg: %s", err)
		}

		// Очищаем состояние пользователя после завершения процесса
		delete(UserStates, chatId)
		delete(UserRtsp, chatId)

		if err := bot.MakeNewMain(chatId); err != nil {
			return fmt.Errorf("Can not make new main: %s", err)
		}
	} else if UserStates[chatId] == "waitingForNameOfCamera" {
		cameraName := msg.Text
		s := CameraName[chatId]

		camera, err := bot.camerasAPI.GetCameraByName(s)
		if err != nil {
			return fmt.Errorf("Can not get camera by name: %s", err)
		}

		err = bot.camerasAPI.UpdateCamera(models.Camera{
			Name: cameraName,
			Rtsp: camera.Rtsp,
		})
		if err != nil {
			return fmt.Errorf("Can not update camera name: %s", err)
		}

		// Подтверждаем сохранение
		err = bot.SendMessage(
			chatId,
			fmt.Sprintf("Камера успешно обновлена!\nНазвание: %s\nRTSP-адрес: %s\n",
				cameraName,
				camera.Rtsp),
			nil,
		)
		if err != nil {
			return fmt.Errorf("Can not send msg: %s", err)
		}

		// Очищаем состояние пользователя после завершения процесса
		delete(UserStates, chatId)
		delete(UserRtsp, chatId)
		delete(CameraName, chatId)

		if err := bot.MakeNewMain(chatId); err != nil {
			return fmt.Errorf("Can not make new main: %s", err)
		}
	} else {
		msg := tg.NewMessage(msg.Chat.ID, "Я не знаю такую команду :(\nЧтобы начать мной пользоваться, отправьте мне слово \"Начать\"")
		if _, err := bot.tgAPI.Send(msg); err != nil {
			return fmt.Errorf("Can not send msg: %s", err)
		}
	}

	return nil
}

func (bot *Bot) MakeMessage(update tg.Update, desc string) error {
	msg := tg.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, desc)

	if _, err := bot.tgAPI.Send(msg); err != nil {
		return fmt.Errorf("Can not send msg: %w", err)
	}

	return nil
}

func (bot *Bot) NotifyReady(notify models.Notify) error {
	parts := strings.Split(notify.FilePath, "/")
	data, err := bot.camerasAPI.GetFile(parts[1], parts[2])
	if err != nil {
		return fmt.Errorf("Invalid get file: %w", err)
	}

	reader := bytes.NewReader(data)

	doc := tg.FileReader{
		Name:   notify.FilePath,
		Reader: reader,
	}

	if notify.Format == "image" {
		msg := tg.NewPhoto(notify.ChatID, doc)
		_, err = bot.tgAPI.Send(msg)
		if err != nil {
			return fmt.Errorf("Invalid send file: %w", err)
		}
	} else {
		msg := tg.NewVideo(notify.ChatID, doc)
		_, err = bot.tgAPI.Send(msg)
		if err != nil {
			return fmt.Errorf("Invalid send file: %w", err)
		}

	}

	return nil
}

func (bot *Bot) NotifyAlert(fileName string) error {
	data, err := bot.camerasAPI.GetFile("alert", fileName)
	if err != nil {
		return fmt.Errorf("Invalid get file: %w", err)
	}

	filePath := fmt.Sprintf("alert/%s.png", fileName)

	users, err := bot.usersAPI.GetAllUsers()
	if err != nil {
		return fmt.Errorf("Invalid get all users: %w", err)
	}

	for _, user := range users {
		reader := bytes.NewReader(data)

		doc := tg.FileReader{
			Name:   filePath,
			Reader: reader,
		}

		fileNameInt, err := strconv.ParseInt(fileName[:len(fileName)-4], 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid parse file name: %w", err)
		}

		t := time.Unix(fileNameInt, 0)
		formatted := t.Format("02.01.2006 15:04:05")
		caption := fmt.Sprintf("🚨 Движение зафиксировано в %s", formatted)

		msg := tg.NewPhoto(user.Chat_ID, doc)
		msg.Caption = caption

		_, err = bot.tgAPI.Send(msg)
		if err != nil {
			bot.log.Error().Msgf("Invalid send file: %s to user %d", err, user.Chat_ID)
		}
	}

	return nil
}

func (bot *Bot) MakeNewMain(chatId int64) error {
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

	err := bot.SendMessage(
		chatId,
		menu,
		&buttons,
	)
	if err != nil {
		return fmt.Errorf("Can not send msg: %s", err)
	}

	return nil
}
