package api

import (
	"fmt"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

type UsersAPI struct {
	client *resty.Client
}

func NewUsersApi(url string) *UsersAPI {
	return &UsersAPI{
		client: resty.New().SetBaseURL(url).SetTimeout(1*time.Minute).SetBasicAuth("dev", "test"),
	}
}

func (api *UsersAPI) GetAllUsers() ([]models.User, error) {
	const endpoint = "/users/get"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		return nil, err
	}

	resp.Body()
	allUsers := []models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &allUsers); err != nil {
		return nil, err
	}

	return allUsers, nil
}

func (api *UsersAPI) AddUser(user models.User) (bool, error) {
	const endpoint = "/users/add"

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

func (api *UsersAPI) GetUserByChatID(chatId int64) (models.User, error) {
	endpoint := "/users/get-%d"

	resp, err := api.client.R().SetBasicAuth("dev", "test").Get(fmt.Sprintf(endpoint, chatId))
	if err != nil {
		return models.User{}, err
	}

	user := models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (api *UsersAPI) GetAdmin() (models.User, error) {
	const endpoint = "/users/getAdmin"

	resp, err := api.client.R().SetBasicAuth("dev", "test").Get(endpoint)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &user); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (api *UsersAPI) UpdateUser(updatedUser models.User) (bool, error) {
	const endpoint = "/users/update"

	resp, err := api.client.R().SetBody(updatedUser).SetBasicAuth("dev", "test").Put(endpoint)
	if err != nil {
		return false, err
	}

	return resp.StatusCode() == 204, nil
}
