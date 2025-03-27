package api

import (
	"fmt"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/rtsp_multi/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/rtsp_multi/internal/models"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

type CameraAPI struct {
	client *resty.Client
}

func NewCameraApi(cfg config.Config) *CameraAPI {
	return &CameraAPI{
		client: resty.New().SetBaseURL(cfg.CamerasUrl).
			SetTimeout(1 * time.Minute),
	}
}

func (api *CameraAPI) GetAllCameras() ([]models.Camera, error) {
	const endpoint = "/cameras/get"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		return nil, err
	}

	allCameras := []models.Camera{}
	if err := jsoniter.Unmarshal(resp.Body(), &allCameras); err != nil {
		return nil, err
	}

	return allCameras, nil
}

func (api *CameraAPI) GetCameraByName(name string) (bool, error) {
	endpoint := "/cameras/get-%s"

	resp, err := api.client.R().Get(fmt.Sprintf(endpoint, name))
	if err != nil {
		return false, err
	}

	camera := models.Camera{}
	if err := jsoniter.Unmarshal(resp.Body(), &camera); err != nil {
		return false, err
	}

	return true, nil
}
