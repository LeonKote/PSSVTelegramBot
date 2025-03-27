package api

import (
	"fmt"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/models"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

type UsersAPI struct {
	client *resty.Client
}

func NewUsersApi(cfg config.Config) *UsersAPI {
	return &UsersAPI{
		client: resty.New().SetBaseURL(cfg.UsersApi).
			SetTimeout(1 * time.Minute),
	}
}

func (api *UsersAPI) GetAllUsers() ([]models.User, error) {
	const endpoint = "/users/get"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Can not get all users: %w", err)
	}

	allUsers := []models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &allUsers); err != nil {
		return nil, fmt.Errorf("Can not unmarshal all users: %w", err)
	}

	return allUsers, nil
}

func (api *UsersAPI) AddUser(user models.User) (bool, error) {
	const endpoint = "/users/add"

	resp, err := api.client.R().SetBody(user).Post(endpoint)
	if err != nil {
		return false, fmt.Errorf("Can not add user: %w", err)
	}

	if resp.StatusCode() != 204 {
		return false, fmt.Errorf("Invalid status code: %d", resp.StatusCode())
	}

	return true, nil
}

func (api *UsersAPI) GetUserByChatID(chatId int64) (models.User, error) {
	endpoint := "/users/get-%d"

	a := fmt.Sprintf(endpoint, chatId)
	resp, err := api.client.R().Get(a)
	if err != nil {
		return models.User{}, fmt.Errorf("Can not get user by chat id: %w", err)
	}

	if resp.StatusCode() == 204 {
		return models.User{}, nil
	}

	user := models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &user); err != nil {
		return models.User{}, fmt.Errorf("Can not unmarshal user by chat id: %w", err)
	}

	return user, nil
}

func (api *UsersAPI) GetAdmin() (models.User, error) {
	const endpoint = "/users/getAdmin"

	resp, err := api.client.R().Get(endpoint)
	if err != nil {
		return models.User{}, fmt.Errorf("Can not get admin: %w", err)
	}

	user := models.User{}
	if err := jsoniter.Unmarshal(resp.Body(), &user); err != nil {
		return models.User{}, fmt.Errorf("Can not unmarshal admin: %w", err)
	}

	return user, nil
}

func (api *UsersAPI) UpdateUser(updatedUser models.User) (bool, error) {
	const endpoint = "/users/update"

	resp, err := api.client.R().SetBody(updatedUser).Put(endpoint)
	if err != nil {
		return false, fmt.Errorf("Can not update user: %w", err)
	}

	if resp.StatusCode() != 204 {
		return false, fmt.Errorf("Invalid status code: %d", resp.StatusCode())
	}

	return true, nil
}
