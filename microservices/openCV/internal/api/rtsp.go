package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Impisigmatus/service_core/log"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Camera struct {
	stream *ffmpeg.Stream
}

func NewCamera(rtspURL string) *Camera {
	return &Camera{
		stream: ffmpeg.Input(rtspURL),
	}
}

// chat_id
// file(photo/video)

func (cam *Camera) RecordVideo(ctx context.Context, duration int, sourceStreamURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceStreamURL, nil)
	if err != nil {
		log.Errorf("Error creating request: %s", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Error connecting to camera: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	var buffer bytes.Buffer

	// Просто копируем тело до закрытия контекста
	_, err = io.Copy(&buffer, resp.Body)
	if err != nil {
		// Игнорируем таймаут как нормальное завершение
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
			log.Info("Видео обрезано по времени — без ошибок.")
		} else {
			log.Errorf("Ошибка при копировании: %s", err)
			return nil, err
		}
	}

	return io.NopCloser(&buffer), nil
}
