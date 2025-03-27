package api

import (
	"fmt"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

type CameraAPI struct {
	client *resty.Client
}

func NewCameraApi(cfg config.Config) *CameraAPI {
	return &CameraAPI{
		client: resty.New().SetBaseURL(cfg.CamerasApi).
			SetTimeout(1 * time.Minute),
	}
}

func (api *CameraAPI) GetAllCameras() ([]models.Camera, error) {
	const endpoint = "/cameras/get"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Can not get all cameras: %w", err)
	}

	allCameras := []models.Camera{}
	if err := jsoniter.Unmarshal(resp.Body(), &allCameras); err != nil {
		return nil, fmt.Errorf("Can not unmarshal all cameras: %w", err)
	}

	return allCameras, nil
}

func (api *CameraAPI) AddCamera(user models.User) (bool, error) {
	const endpoint = "/cameras/add"

	resp, err := api.client.R().SetBody(user).Post(endpoint)
	if err != nil {
		return false, fmt.Errorf("Can not add camera: %w", err)
	}

	allUsers := []models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &allUsers); err != nil {
		return false, fmt.Errorf("Can not unmarshal all cameras: %w", err)
	}

	return true, nil
}

func (api *CameraAPI) GetCameraByName(name string) (models.Camera, error) {
	endpoint := "/cameras/get-%s"

	resp, err := api.client.R().Get(fmt.Sprintf(endpoint, name))
	if err != nil {
		return models.Camera{}, fmt.Errorf("Can not get camera by name: %w", err)
	}

	camera := models.Camera{}
	if err := jsoniter.Unmarshal(resp.Body(), &camera); err != nil {
		return models.Camera{}, fmt.Errorf("Can not unmarshal camera by name: %w", err)
	}

	return camera, nil
}

func (api *CameraAPI) Capture(tailEndpoint string, record models.Record) error {
	endpoint := "/cameras/%s"

	if record.Duration == nil {
		endpoint = fmt.Sprintf(endpoint, tailEndpoint)
	} else {
		endpoint = fmt.Sprintf(endpoint, tailEndpoint)
	}

	resp, err := api.client.R().SetBody(record).Post(endpoint)
	if err != nil {
		return fmt.Errorf("Can not make video: %w", err)
	}

	uuid := models.Uuid{}
	if err := jsoniter.Unmarshal(resp.Body(), &uuid); err != nil {
		return fmt.Errorf("Can not unmarshal video: %w", err)
	}

	return nil
}

func (api *CameraAPI) GetFile(chatId string, uuid string) ([]byte, error) {
	const endpoint = "/cameras/%s/%s/get"

	resp, err := api.client.R().Get(fmt.Sprintf(endpoint, chatId, uuid))
	if err != nil {
		return nil, fmt.Errorf("Can not get file: %w", err)
	}

	return resp.Body(), nil
}
