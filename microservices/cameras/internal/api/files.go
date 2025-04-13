package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
)

type FilesAPI struct {
	client *resty.Client
}

func NewFilesApi(cfg config.Config) *FilesAPI {
	auth := strings.Split(cfg.BasicAuth, ":")
	return &FilesAPI{
		client: resty.New().
			SetBaseURL(cfg.FilesHost).
			SetTimeout(1*time.Minute).
			SetBasicAuth(auth[0], auth[1]),
	}
}

func (api *FilesAPI) GetAllFiles(logger zerolog.Logger) ([]models.File, error) {
	const endpoint = "/files/get"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		logger.Error().Msgf("Error creating request: %s", err)
		return nil, err
	}

	allFiles := []models.File{}
	if err := jsoniter.Unmarshal(resp.Body(), &allFiles); err != nil {
		logger.Error().Msgf("Error unmarshalling response: %s", err)
		return nil, err
	}

	return allFiles, nil
}

func (api *FilesAPI) AddFile(logger zerolog.Logger, file models.File) (bool, error) {
	const endpoint = "/files/add"

	resp, err := api.client.R().SetBody(file).Post(endpoint)
	if err != nil {
		logger.Error().Msgf("Error creating request: %s", err)
		return false, err
	}

	if resp.StatusCode() != 204 {
		logger.Error().Msgf("Invalid status code: %d", resp.StatusCode())
		return false, err
	}

	return true, nil
}

func (api *FilesAPI) GetFileByFilePath(logger zerolog.Logger, uuid string) (bool, error) {
	endpoint := "/files/get-%s"

	resp, err := api.client.R().Get(fmt.Sprintf(endpoint, uuid))
	if err != nil {
		logger.Error().Msgf("Error creating request: %s", err)
		return false, err
	}

	file := models.File{}
	if err := jsoniter.Unmarshal(resp.Body(), &file); err != nil {
		logger.Error().Msgf("Error unmarshalling response: %s", err)
		return false, err
	}

	return true, nil
}

func (api *FilesAPI) ChangeStatus(logger zerolog.Logger, uuid string, fileSize int, status string) error {
	endpoint := "/files%s/%d/%s"
	a := fmt.Sprintf(endpoint, uuid, fileSize, status)
	fmt.Println(a)

	resp, err := api.client.R().Put(a)
	if err != nil {
		logger.Error().Msgf("Error creating request: %s", err)
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		logger.Error().Msgf("Invalid status code: %d", resp.StatusCode())
		return err
	}

	return nil
}
