package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Impisigmatus/service_core/utils"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/app"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/models"
	jsoniter "github.com/json-iterator/go"
)

type Transport struct {
	app *app.Application
}

func NewTransport(app *app.Application) server.ServerInterface {
	return &Transport{
		app: app,
	}
}

// Set godoc
//
// @Router /api/files/add [post]
// @Summary Добавление информации о файле в БД
// @Description При обращении, добавляется информация о файле
//
// @Tags APIs
// @Accept       application/json
// @Produce      application/json
// @Param 	request	body	file	true	"Тело запроса"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PostApiFilesAdd(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid read body: %s", err), "Не удалось прочитать тело запроса")
		return
	}

	var file models.File
	if err := jsoniter.Unmarshal(data, &file); err != nil {
		utils.WriteString(w, http.StatusBadRequest, fmt.Errorf("Invalid parse body: %s", err), "Не удалось распарсить тело запроса формата JSON")
		return
	}

	if err := transport.app.AddFile(file); err != nil {
		utils.WriteString(w, http.StatusInternalServerError, err, "Не удалось добавить информацию о файле")
		return
	}

	utils.WriteNoContent(w)
}

// Set godoc
//
// @Router /api/files/delete-{uuid} [delete]
// @Summary Удаление файла из БД
// @Description При обращении, удаляет файл из БД по его uuid
//
// @Tags APIs
// @Produce      application/json
//
// @Param	uuid	path	string	true	"UUID файла"
//
// @Success 200 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизац®ии"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) DeleteApiFilesDeleteUuid(w http.ResponseWriter, r *http.Request, uuid string) {
	err := transport.app.RemoveFile(uuid)
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid remove file: %s", err), "Невозможно удалить файл")
		return
	}

	utils.WriteString(w, http.StatusOK, nil, "Файла удалён")
}

// Set godoc
//
// @Router /api/files/get-{uuid} [get]
// @Summary Получение файла по его uuid
// @Description При обращении, возвращает юзера по его uuid
//
// @Tags APIs
// @Produce      application/json
//
// @Param	uuid	path	string	true	"UUID файла"
//
// @Success 200 {object} file "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiFilesGetUuid(w http.ResponseWriter, r *http.Request, uuid string) {
	file, err := transport.app.GetByUuid(uuid)
	if err != nil {
		utils.WriteString(w, http.StatusBadRequest, fmt.Errorf("Invalid get file: %s", err), "Не удалось получить файл")
		return
	}

	utils.WriteObject(w, file)
}

// Set godoc
//
// @Router /api/files/get [get]
// @Summary Получение всех юзеров
// @Description При обращении, возвращает всех юзеров
//
// @Tags APIs
// @Produce      application/json
//
// @Success 200 {array} file "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 404 {object} nil "Ошибка получения данных"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) GetApiFilesGet(w http.ResponseWriter, r *http.Request) {
	files, err := transport.app.GetAllFiles()
	if err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid get files: %s", err), "Не удалось получить информацию о файлах")
		return
	}
	if len(files) == 0 {
		utils.WriteString(w, http.StatusNotFound, fmt.Errorf("DB is empty: %s", err), "В базе нет файлов")
		return
	}

	utils.WriteObject(w, files)
}

// Set godoc
//
// @Router /api/files/{uuid}/{fileName}/{file_size}/{status} [put]
// @Summary Обновление данных о файле
// @Description При обращении, обновляет данные
//
// @Tags APIs
// @Accept       application/json
// @Produce      application/json
//
// @Param	uuid	path	string	true	"UUID файла"
// @Param	fileName	path	string	true	"Имя файла"
// @Param	file_size	path	int	true	"Размер файла"
// @Param	status	path	string	true	"Статус файла"
//
// @Success 204 {object} nil "Запрос выполнен успешно"
// @Failure 400 {object} nil "Ошибка валидации данных"
// @Failure 401 {object} nil "Ошибка авторизации"
// @Failure 404 {object} nil "Ошибка получения данных"
// @Failure 500 {object} nil "Произошла внутренняя ошибка сервера"
func (transport *Transport) PutApiFilesUuidFileNameFileSizeStatus(w http.ResponseWriter, r *http.Request, uuid string, fileName string, fileSize int, status string) {
	filePath := fmt.Sprintf("/%s/%s", uuid, fileName)
	if err := transport.app.UpdateFile(filePath, status, fileSize); err != nil {
		utils.WriteString(w, http.StatusInternalServerError, fmt.Errorf("Invalid update file: %s", err), "Не удалось обновить файл")
	}

	utils.WriteNoContent(w)
}
