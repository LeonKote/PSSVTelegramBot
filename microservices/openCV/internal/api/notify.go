package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type NotifyAPI struct {
	client *resty.Client
}

func NewNotifyAPI(url string) *NotifyAPI {
	return &NotifyAPI{
		client: resty.New().SetBaseURL(url).
			SetTimeout(1 * time.Minute),
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
