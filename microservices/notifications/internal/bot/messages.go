package bot

import (
	"fmt"
	"log"
	"strconv"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	logrus "github.com/sirupsen/logrus"
)

func (bot *Bot) HandleMessage(msg *tg.Message) {
	if msg.Text == "/start" || msg.Text == "Начать" {
		user, err := bot.usersAPI.GetUserByChatID(msg.Chat.ID)
		if err != nil {
			logrus.Errorf("Invalid get user by chat id: %s", err)
		}

		if msg.Chat.ID == bot.admin && user == (models.User{}) {
			ok, err := bot.usersAPI.AddUser(models.User{
				Chat_ID:  msg.Chat.ID,
				Username: msg.From.UserName,
				Name:     msg.From.FirstName,
				Is_Admin: true,
				Status:   "approved",
			})
			if err != nil || !ok {
				logrus.Errorf("Invalid add user: %s", err)
			}
		}

		switch user.Status {
		case "approved":
			buttons := tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData("Начать", toMain),
				),
			)

			msg := tg.NewMessage(msg.Chat.ID, "Добро пожаловать! Нажмите кнопку, чтобы начать испольование бота.")
			msg.ReplyMarkup = buttons

			if _, err := bot.tgAPI.Send(msg); err != nil {
				log.Printf("Can not send msg: %S", err)
			}
		case "pending":
			msg := tg.NewMessage(msg.Chat.ID, "Ожидает подтверждения.")
			if _, err := bot.tgAPI.Send(msg); err != nil {
				log.Printf("Can not send msg: %S", err)
			}
		case "rejected":
			msg := tg.NewMessage(msg.Chat.ID, "❌ Вам отказано в доступе.")
			if _, err := bot.tgAPI.Send(msg); err != nil {
				log.Printf("Can not send msg: %S", err)
			}

		}

		if user == (models.User{}) {
			ok, err := bot.usersAPI.AddUser(models.User{
				Chat_ID:  msg.Chat.ID,
				Username: msg.From.UserName,
				Name:     msg.From.FirstName,
				Is_Admin: false,
				Status:   "pending",
			})
			if err != nil || !ok {
				logrus.Errorf("Invalid add user: %s", err)
			}

			msgToAdmin := tg.NewMessage(bot.admin,
				fmt.Sprintf("❗ Новый пользователь запросил доступ:\n\n👤 %s (@%s)\n🔢 ID: %d",
					fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName),
					msg.From.UserName,
					msg.From.ID))

			keyboard := tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData("✅ Добавить", "approve:"+strconv.FormatInt(msg.From.ID, 10)),
					tg.NewInlineKeyboardButtonData("❌ Отклонить", "reject:"+strconv.FormatInt(msg.From.ID, 10)),
				),
			)

			msgToAdmin.ReplyMarkup = keyboard
			bot.tgAPI.Send(msgToAdmin)

			// Уведомляем пользователя
			newMsg, err := bot.tgAPI.Send(tg.NewMessage(msg.From.ID, "Ваш запрос отправлен администратору на проверку. Ожидайте."))
			if err != nil {
				log.Printf("Can not send msg: %S", err)
			}
			LastActions[msg.From.ID] = newMsg.MessageID
		}
	} else {
		msg := tg.NewMessage(msg.Chat.ID, "Я не знаю такую команду :(\nЧтобы начать мной пользоваться, отправьте мне слово \"Начать\"")
		if _, err := bot.tgAPI.Send(msg); err != nil {
			log.Printf("Can not send msg: %S", err)
		}
	}
}

func (bot *Bot) MakeMessage(update tg.Update, desc string) {
	msg := tg.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, desc)

	if _, err := bot.tgAPI.Send(msg); err != nil {
		log.Printf("Can not send msg: %S", err)
	}
}
