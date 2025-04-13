package service

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Impisigmatus/service_core/log"
	"github.com/Impisigmatus/service_core/utils"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/internal/models"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/internal/repository"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
)

type Transport struct {
	repo repository.IUsersRepository
}

func NewTransport(repo repository.IUsersRepository) server.ServerInterface {
	return &Transport{
		repo: repo,
	}
}

// Set godoc
//
// @Router /api/users/add [post]
// @Summary Добавление юзера в БД
// @Description При обращении, добавляется отклик в БД по телу запроса
//
// @Tags APIs
// @Accept       application/json
// @Produce      application/json
// @Param 	request	body	user	true	"Тело запроса"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiUsersAdd(w http.ResponseWriter, r *http.Request) {
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

	var user models.User
	if err := jsoniter.Unmarshal(data, &user); err != nil {
		utils.WriteString(log, w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
		return
	}

	if err := transport.repo.AddUser(user); err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, err, "Не удалось добавить пользователя")
		return
	}

	utils.WriteNoContent(log, w)
}

// Set godoc
//
// @Router /api/users/delete-{chat_id} [delete]
// @Summary Удаление юзера из БД
// @Description При обращении, удаляет юзера из БД по его chat_id
//
// @Tags APIs
// @Produce      application/json
// @Param	chat_id	path	int	true	"Chat_id пользователя"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) DeleteApiUsersDeleteChatId(w http.ResponseWriter, r *http.Request, chatId int) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	ok, err := transport.repo.RemoveUser(int64(chatId))
	if err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, err, "Пользователя не существует")
		return
	}

	if ok {
		utils.WriteString(log, w, http.StatusOK, nil, "Пользователя не существует")
		return
	} else {
		utils.WriteNoContent(log, w)
		return
	}
}

// Set godoc
//
// @Router /api/users/get-{chat_id} [get]
// @Summary Получение юзера по его chat_id
// @Description При обращении, возвращает юзера по его chat_id
//
// @Tags APIs
// @Produce      application/json
// @Param	chat_id	path	int	true	"Chat_id пользователя"
//
// @Success 200 {object} user "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiUsersGetChatId(w http.ResponseWriter, r *http.Request, chatId int) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	user, err := transport.repo.GetUserByChatID(int64(chatId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteNoContent(log, w)
			return
		}

		utils.WriteString(log, w, http.StatusInternalServerError, err, "Не удалось получить пользователя")
		return
	}

	utils.WriteObject(log, w, user)
}

// Set godoc
//
// @Router /api/users/get [get]
// @Summary Получение всех юзеров
// @Description При обращении, возвращает всех юзеров
//
// @Tags APIs
// @Produce      application/json
//
// @Success 200 {array} user "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiUsersGet(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	users, err := transport.repo.GetAllUsers()
	if err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, err, "Не удалось получить пользователей")
		return
	}
	if len(users) == 0 {
		utils.WriteString(log, w, http.StatusInternalServerError, err, "В базе нет пользователей")
		return
	}

	utils.WriteObject(log, w, users)
}

// Set godoc
//
// @Router /api/users/getAdmin [get]
// @Summary Получение админа
// @Description При обращении, возвращает данные админа
//
// @Tags APIs
// @Produce      application/json
//
// @Success 200 {array} user "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiUsersGetAdmin(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(log.CtxKey).(zerolog.Logger)
	if !ok {
		utils.WriteString(zerolog.Logger{}, w, http.StatusInternalServerError, fmt.Errorf("Invalid logger"), "Невалидный логгер")
		return
	}

	users, err := transport.repo.GetAdminUser()
	if err != nil {
		if users == (models.User{}) {
			utils.WriteString(log, w, http.StatusOK, nil, "Пользователя не существует")
			return
		}

		utils.WriteString(log, w, http.StatusNoContent, err, "Не удалось получить пользователя")
		return
	}

	utils.WriteObject(log, w, users)
}

// Set godoc
//
// @Router /api/users/update [put]
// @Summary Обновление данных пользователя
// @Description При обращении, обновляет данные пользоватя
//
// @Tags APIs
// @Accept       application/json
// @Produce      application/json
// @Param 	request	body	user	true	"Тело запроса"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PutApiUsersUpdate(w http.ResponseWriter, r *http.Request) {
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

	var updateUser models.User
	if err := jsoniter.Unmarshal(data, &updateUser); err != nil {
		utils.WriteString(log, w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
		return
	}

	ok, err = transport.repo.UpdateUser(updateUser)
	if err != nil {
		utils.WriteString(log, w, http.StatusInternalServerError, err, "Не удалось обновить данные пользователя")
		return
	}

	if ok {
		utils.WriteNoContent(log, w)
		return
	} else {

		utils.WriteString(log, w, http.StatusOK, nil, "Пользователя не существует")
		return
	}
}
