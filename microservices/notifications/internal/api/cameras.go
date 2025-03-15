package api

import (
	"fmt"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

type CameraAPI struct {
	client *resty.Client
}

func NewCameraApi(url string) *CameraAPI {
	return &CameraAPI{
		client: resty.New().SetBaseURL(url).SetTimeout(1*time.Minute).SetBasicAuth("dev", "test"),
	}
}

func (api *CameraAPI) GetAllCameras() ([]models.Camera, error) {
	const endpoint = "/cameras/get"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		return nil, err
	}

	resp.Body()
	allCameras := []models.Camera{}
	if err := jsoniter.Unmarshal(resp.Body(), &allCameras); err != nil {
		return nil, err
	}

	return allCameras, nil
}

func (api *CameraAPI) AddCamera(user models.User) (bool, error) {
	const endpoint = "/cameras/add"

	resp, err := api.client.R().SetBody(user).Post(endpoint)
	if err != nil {
		return false, err
	}

	allUsers := []models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &allUsers); err != nil {
		return false, err
	}

	return true, nil
}

func (api *CameraAPI) GetCameraByID(id int) (bool, error) {
	endpoint := "/cameras/get-%d"

	resp, err := api.client.R().SetBasicAuth("dev", "test").Get(fmt.Sprintf(endpoint, id))
	if err != nil {
		return false, err
	}

	user := models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &user); err != nil {
		return false, err
	}

	return true, nil
}
