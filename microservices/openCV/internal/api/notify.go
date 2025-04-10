package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LeonKote/PSSVTelegramBot/microservices/openCV/internal/config"
	"github.com/go-resty/resty/v2"
)

type NotifyAPI struct {
	client *resty.Client
}

func NewNotifyAPI(cfg config.Config) *NotifyAPI {
	auth := strings.Split(cfg.BasicAuth, ":")
	return &NotifyAPI{
		client: resty.New().SetBaseURL(cfg.NotificationsApi).
			SetTimeout(1*time.Minute).
			SetBasicAuth(auth[0], auth[1]),
	}
}

func (api *NotifyAPI) SendAlert(filePath string) (bool, error) {
	const endpoint = "/notify/%s"

	resp, err := api.client.R().Post(fmt.Sprintf(endpoint, filePath))
	if err != nil {
		return false, fmt.Errorf("Can not send alert: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return false, fmt.Errorf("Can not send alert: %s", resp.Status())
	}

	return true, nil
}
