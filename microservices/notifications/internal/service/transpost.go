package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Impisigmatus/service_core/log"
	"github.com/Impisigmatus/service_core/utils"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/bot"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
)

type Transport struct {
	bot *bot.Bot
}

func NewTransport(bot *bot.Bot) server.ServerInterface {
	go bot.Run()

	return &Transport{
		bot: bot,
	}
}

// Set godoc
//
// @Router /api/notify [post]
// @Summary Отправка оповещения
// @Description При обращении, отправляет оповещение
//
// @Tags APIs
// @Accept       application/json
//
// @Param	request	body	queue	true	"Тело запроса"
//
// @Success 204 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 404 {object} nil "Ошибка получения данных"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiNotify(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, fmt.Errorf("Invalid read body: %s", err), "Не удалось прочитать тело запроса")
		return
	}

	var notification models.Notify
	if err := jsoniter.Unmarshal(data, &notification); err != nil {
		utils.WriteString(log, w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
		return
	}

	if err := transport.bot.NotifyReady(notification); err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, fmt.Errorf("Invalid notify: %s", err), "Не удалось отправить оповещения")
		return
	}

	utils.WriteNoContent(log, w)
}

// Set godoc
//
// @Router /api/notify/alert/{file_name} [post]
// @Summary Отправка оповещения при alert'ах
// @Description При обращении, отправляет оповещение при alert'ах
//
// @Tags APIs
// @Accept       application/json
//
// @Param	file_name	path	string	true	"Название файла"
//
// @Success 204 {object} nil "Запрос выполнен успешно"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 404 {object} nil "Ошибка получения данных"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiNotifyAlertFileName(w http.ResponseWriter, r *http.Request, fileName string) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	if err := transport.bot.NotifyAlert(fileName); err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, fmt.Errorf("Invalid notify: %s", err), "Не удалось отправить оповещения")
		return
	}

	utils.WriteNoContent(log, w)
}
