package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) HandleCallback(update tg.Update) {
	data := update.CallbackData()

	// Разбираем данные кнопки ("approve:123456789")
	action, userID, err := bot.parseCallbackData(data)
	if err != nil {
		log.Printf("Can not parse callback data: %s", err)
	}

	switch action {
	case "approve":
		user, err := bot.usersAPI.GetUserByChatID(userID)
		if err != nil {
			log.Printf("Can not get user by chat_id: %s", err)
		}

		ok, err := bot.usersAPI.UpdateUser(models.User{
			Chat_ID:  userID,
			Username: user.Username,
			Name:     user.Name,
			Is_Admin: false,
			Status:   "approved",
		})
		if err != nil || !ok {
			log.Printf("Can not update user: %s", err)
		}

		messageID, exists := LastActions[userID]
		if !exists {
			log.Printf("Сообщение пользователя не найдено")
			return
		}

		editMsg := tg.NewEditMessageText(userID, messageID, "🎉 Вам одобрили доступ!")
		newMsg, err := bot.tgAPI.Send(editMsg)
		if err != nil {
			log.Printf("Can not send msg: %s", err)
		} else {
			LastActions[userID] = newMsg.MessageID
		}

		bot.MakeMessage(update, "✅ Пользователь одобрен!")

		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Начать", toMain),
			),
		)

		msg := tg.NewMessage(userID, "✅ Вам разрешили использование бота! Нажмите кнопку \"Начать\", чтобы начать испольование бота.")
		msg.ReplyMarkup = buttons

		if _, err := bot.tgAPI.Send(msg); err != nil {
			log.Printf("Can not send msg: %s", err)
		}

	case "reject":
		user, err := bot.usersAPI.GetUserByChatID(userID)
		if err != nil {
			log.Printf("Can not get user by chat_id: %s", err)
		}

		ok, err := bot.usersAPI.UpdateUser(models.User{
			Chat_ID:  userID,
			Username: user.Username,
			Name:     user.Name,
			Is_Admin: false,
			Status:   "rejected",
		})
		if err != nil || !ok {
			log.Printf("Can not update user: %s", err)
		}

		// Сохранение ID последнего сообщения от пользователя
		messageID, exists := LastActions[userID]
		if !exists {
			log.Println("Ошибка: сообщение пользователя не найдено")
			return
		}

		editMsg := tg.NewEditMessageText(userID, messageID, "🚫 Ваш запрос отклонён.")
		newMsg, err := bot.tgAPI.Send(editMsg)
		if err != nil {
			log.Println(err)
		} else {
			LastActions[userID] = newMsg.MessageID
		}

		bot.MakeMessage(update, "❌ Пользователь отклонён.")
	case toCameras:
		cameras, err := bot.camerasAPI.GetAllCameras()
		if err != nil {
			log.Printf("Invalid get all cameras: %s", err)
		}

		camerasMarkup := bot.getCameraButtons(cameras)
		bot.MakeButton(update, "📸 Список камер:", camerasMarkup)
	case toUsers:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Назад", toMain),
			),
		)

		bot.MakeButton(update, "Добавить камеру автоматически", buttons)
	case toAdd:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Назад", toMain),
			),
		)

		bot.MakeButton(update, "Добавить камеру по её MAC-адресу", buttons)
	case toMain:
		buttons := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("📸 Список камер", toCameras),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Добавить камеру автоматически", toUsers),
			),
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("Добавить камеру по её MAC-адресу", toAdd),
			),
		)
		bot.MakeButton(update, "Меню", buttons)
	}
}

func (bot *Bot) MakeButton(update tg.Update, desc string, buttons tg.InlineKeyboardMarkup) {
	msg := tg.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, desc)

	keyboard := buttons
	msg.ReplyMarkup = &keyboard

	if _, err := bot.tgAPI.Send(msg); err != nil {
		log.Println(err)
	}
}

func (bot *Bot) parseCallbackData(data string) (action string, userID int64, err error) {
	parts := strings.Split(data, ":")

	action = parts[0] // ✅ Первое значение всегда есть

	if len(parts) > 1 { // ✅ Проверяем, есть ли второй параметр
		userID, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", 0, fmt.Errorf("Ошибка парсинга userID: %w", err)
		}
	} else {
		userID = 0 // ✅ Если userID нет, используем 0
	}

	return action, userID, nil
}

func (bot *Bot) getCameraButtons(cameras []models.Camera) tg.InlineKeyboardMarkup {
	var buttons [][]tg.InlineKeyboardButton

	for index, camera := range cameras {
		if camera.Name == "" {
			camera.Name = fmt.Sprintf("Камера %d", index)
		}
		// Создаём кнопку для каждой камеры
		button := tg.NewInlineKeyboardButtonData(camera.Name, fmt.Sprintf("camera:%d", camera.Id))
		buttons = append(buttons, tg.NewInlineKeyboardRow(button))
	}
	back := tg.NewInlineKeyboardButtonData("Назад", toMain)
	buttons = append(buttons, tg.NewInlineKeyboardRow(back))

	return tg.NewInlineKeyboardMarkup(buttons...)
}
