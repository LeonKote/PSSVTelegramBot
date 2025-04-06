package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Impisigmatus/service_core/utils"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/minio/minio-go/v7"
)

type Transport struct {
	app *Application
	ctx context.Context
}

func NewTransport(ctx context.Context, app *Application) server.ServerInterface {
	return &Transport{
		app: app,
		ctx: ctx,
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

	if err := transport.app.repo.AddCamera(camera); err != nil {
		utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось добавить пользователя")
		return
	}

	utils.WriteNoContent(w)
}

// Set godoc
//
// @Router /api/cameras/delete-{name} [delete]
// @Summary Удаление камеры из БД
// @Description При обращении, удаляет камеру из БД по её name
//
// @Tags APIs
// @Produce      application/json
// @Param	name	path	string	true	"Название камеры"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) DeleteApiCamerasDeleteName(w http.ResponseWriter, r *http.Request, name string) {
	ok, err := transport.app.repo.RemoveCamera(name)
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
// @Router /api/cameras/get-{name} [get]
// @Summary Получение камера по её name
// @Description При обращении, возвращает камеру по её name
//
// @Tags APIs
// @Produce      application/json
//
// @Param	name	path	string	true	"Название камеры"
//
// @Success 200 {object} camera "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiCamerasGetName(w http.ResponseWriter, r *http.Request, name string) {
	user, err := transport.app.repo.GetCameraByName(name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteNoContent(w)
			return
		}

		utils.WriteString(w, http.StatusNoContent, err, "Не удалось получить данные о камере")
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
	cameras, err := transport.app.repo.GetAllCameras()
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось получить список камер")
		return
	}
	if len(cameras) == 0 {
		utils.WriteString(w, http.StatusInternalServerError, err, "В базе нет камер")
		return
	}

	utils.WriteObject(w, cameras)
}

// Set godoc
//
// @Router /api/cameras/record [post]
// @Summary Запись видео
// @Description При обращении, записывает видео с камеры, продолжительностью в duration секунд
//
// @Tags APIs
// @Produce      application/json
//
// @Param 	request	body	record	true	"Тело запроса"
//
// @Success 204 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiCamerasRecord(w http.ResponseWriter, r *http.Request) {
	reqId := uuid.New().String()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid read body: %s", err), "Не удалось прочитать тело запроса")
		return
	}

	go func() {
		var record models.Record
		if err := jsoniter.Unmarshal(body, &record); err != nil {
			utils.WriteString(w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
			return
		}

		fileName, err := transport.app.Record(transport.ctx, record, true, reqId)
		if err != nil {
			if err := transport.app.ChangeStatus(fileName, nilFileSize, statusFailed); err != nil {
				utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось добавить запись в очередь")
				return
			}

			utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось записать видео")
			return
		}
	}()

	utils.WriteObject(w, models.Uuid{Uuid: reqId})
}

// Set godoc
//
// @Router /api/cameras/capture [post]
// @Summary Делает скриншот
// @Description При обращении, делает скришноты с камеры
//
// @Tags APIs
// @Produce      application/json
//
// @Param 	request	body	record	true	"Тело запроса"
//
// @Success 204 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiCamerasCapture(w http.ResponseWriter, r *http.Request) {
	reqId := uuid.New().String()
	utils.WriteObject(w, models.Uuid{Uuid: reqId})

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid read body: %s", err), "Не удалось прочитать тело запроса")
		return
	}

	go func() {
		var record models.Record
		if err := jsoniter.Unmarshal(body, &record); err != nil {
			utils.WriteString(w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
			return
		}

		fileName, err := transport.app.Record(transport.ctx, record, false, reqId)
		if err != nil {
			if err := transport.app.ChangeStatus(fileName, nilFileSize, statusFailed); err != nil {
				utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось добавить запись в очередь")
				return
			}

			utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось записать видео")
			return
		}
	}()

	utils.WriteNoContent(w)
}

// Set godoc
//
// @Router /api/cameras/{dir}/{file_name}/get [get]
// @Summary Получение файла
// @Description При обращении, получает файл с s3
//
// @Tags APIs
// @Produce       application/octet-stream
//
// @Param	dir	path	string	true	"Название директории"
// @Param	file_name	path	string	true	"Название файла"
//
// @Success 200 {file} file "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiCamerasDirFileNameGet(w http.ResponseWriter, r *http.Request, dir string, fileName string) {
	path := fmt.Sprintf("/%s/%s", dir, fileName)
	_, err := transport.app.minioClient.StatObject(transport.ctx, transport.app.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			utils.WriteString(w, http.StatusNotFound, fmt.Errorf("File not found: %s", path), "Файл не найден")
			return
		}
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Error checking object existence: %v", err), "Ошибка проверки существования файла")
		return
	}

	object, err := transport.app.minioClient.GetObject(transport.ctx, transport.app.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid read object: %s", err), "Не удалось получить файл")
		return
	}
	defer object.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(object); err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid read object: %s", err), "Не удалось прочитать файл")
		return
	}

	if _, err := io.Copy(w, &buffer); err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid send object: %s", err), "Не удалось прочитать файл")
		return
	}
}
