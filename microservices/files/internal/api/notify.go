package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/models"
	"github.com/go-resty/resty/v2"
)

type INotifyAPI interface {
	Notify(notify models.Notify) error
}

type notifyAPI struct {
	client *resty.Client
}

func NewNotifyApi(cfg config.Config) INotifyAPI {
	return &notifyAPI{
		client: resty.New().
			SetBaseURL(cfg.NotifyHost).
			SetTimeout(1 * time.Minute),
	}
}

func (api *notifyAPI) Notify(notify models.Notify) error {
	const endpoint = "/notify"

	resp, err := api.client.R().SetBody(notify).Post(endpoint)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("Can not send notify")
	}

	return nil
}
