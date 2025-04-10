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
	if msg.Text == "/start" || msg.Text == start {
		user, err := bot.usersAPI.GetUserByChatID(msg.Chat.ID)
		if err != nil {
			return fmt.Errorf("Invalid get user by chat id: %s", err)
		}

		if msg.Chat.ID == bot.admin && user == (models.User{}) {
			ok, err := bot.usersAPI.AddUser(models.User{
				Chat_ID:  msg.Chat.ID,
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
				Chat_ID:  msg.Chat.ID,
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

			fromChatID := strconv.Itoa(int(msg.From.ID))
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
			newMsg, err := bot.tgAPI.Send(tg.NewMessage(msg.From.ID, "Ваш запрос отправлен администратору на проверку. Ожидайте."))
			if err != nil {
				return fmt.Errorf("Can not send msg: %s", err)
			}
			LastActions[msg.From.ID] = newMsg.MessageID
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
