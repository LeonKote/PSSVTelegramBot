package service

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Impisigmatus/service_core/utils"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

type Transport struct {
	srv *Service
}

func NewTransport(db *sqlx.DB) server.ServerInterface {
	return &Transport{
		srv: NewService(db),
	}
}

// Set godoc
//
// @Router /api/cameras/add [post]
// @Summary Добавление камеры в БД
// @Description При обращении, добавляется камера в БД по телу запроса
//
// @Tags APIs
// @Accept       application/json
// @Produce      application/json
// @Param 	request	body	camera	true	"Тело запроса"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiCamerasAdd(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid read body: %s", err), "Не удалось прочитать тело запроса")
		return
	}

	var camera models.Camera
	if err := jsoniter.Unmarshal(data, &camera); err != nil {
		utils.WriteString(w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
		return
	}

	if err := transport.srv.AddCamera(camera); err != nil {
		utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось добавить пользователя")
		return
	}

	utils.WriteNoContent(w)
}

// Set godoc
//
// @Router /api/cameras/delete-{id} [delete]
// @Summary Удаление камеры из БД
// @Description При обращении, удаляет камеру из БД по её id
//
// @Tags APIs
// @Produce      application/json
// @Param	id	path	int	true	"id камеры"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) DeleteApiCamerasDeleteId(w http.ResponseWriter, r *http.Request, id int) {
	ok, err := transport.srv.RemoveCamera(id)
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, err, "Пользователя не существует")
		return
	}

	if ok {
		utils.WriteString(w, http.StatusOK, nil, "Пользователя не существует")
		return
	} else {
		utils.WriteNoContent(w)
		return
	}
}

// Set godoc
//
// @Router /api/cameras/get-{id} [get]
// @Summary Получение камера по её id
// @Description При обращении, возвращает камеру по её id
//
// @Tags APIs
// @Produce      application/json
// @Param	id	path	int	true	"Chat_id пользователя"
//
// @Success 200 {object} camera "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiCamerasGetId(w http.ResponseWriter, r *http.Request, id int) {
	user, err := transport.srv.GetCameraByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteNoContent(w)
			return
		}

		utils.WriteString(w, http.StatusNoContent, err, "Не удалось получить пользователя")
		return
	}

	utils.WriteObject(w, user)
}

// Set godoc
//
// @Router /api/cameras/get [get]
// @Summary Получение всех камер
// @Description При обращении, возвращает все камеры
//
// @Tags APIs
// @Produce      application/json
//
// @Success 200 {array} camera "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiCamerasGet(w http.ResponseWriter, r *http.Request) {
	users, err := transport.srv.GetAllCameras()
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось получить пользователей")
		return
	}
	if len(users) == 0 {
		utils.WriteString(w, http.StatusInternalServerError, err, "В базе нет пользователей")
		return
	}

	utils.WriteObject(w, users)
}
