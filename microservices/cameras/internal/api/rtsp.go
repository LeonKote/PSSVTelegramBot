package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
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

func (cam *Camera) RecordVideo(logger zerolog.Logger,
	ctx context.Context,
	duration int,
	streamUrl string,
	basicAuth string,
) (io.ReadCloser, error) {
	auth := strings.Split(basicAuth, ":")

	// Контекст с таймаутом на всю запись
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(duration)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, streamUrl, nil)
	if err != nil {
		logger.Error().Msgf("Error creating request: %s", err)
		return nil, err
	}

	req.SetBasicAuth(auth[0], auth[1])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error().Msgf("Error connecting to camera: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	var buffer bytes.Buffer

	// Просто копируем тело до закрытия контекста
	_, err = io.Copy(&buffer, resp.Body)
	if err != nil {
		// Игнорируем таймаут как нормальное завершение
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
			logger.Info().Msg("Видео обрезано по времени — без ошибок.")
		} else {
			logger.Error().Msgf("Ошибка при копировании: %s", err)
			return nil, err
		}
	}

	logger.Info().Msg("Копирование завершено")
	buff, err := cam.encodeVideo(&buffer, duration)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (cam *Camera) CapturePhoto(logger zerolog.Logger, streamUrl string, auth string) (io.ReadCloser, error) {
	logger.Debug().Msgf("Start capture photo at: %s", time.Now())
	var buffer bytes.Buffer
	logger.Debug().Msgf("Url: %s", streamUrl)
	err := ffmpeg.Input(streamUrl, ffmpeg.KwArgs{
		"fflags":  "+genpts", // добавим безопасный флаг
		"f":       "mjpeg",   // явно указываем, что читаем MJPEG
		"headers": auth,
	}).
		Output("pipe:1",
			ffmpeg.KwArgs{
				"f":        "image2pipe", // формат — один файл (image2)
				"c:v":      "mjpeg",
				"frames:v": "1",
			}).
		WithOutput(&buffer).
		WithErrorOutput(os.Stderr).
		Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to encode photo: %w", err)
	}

	logger.Info().Msgf("End capture photo at: %s, len: %d", time.Now(), buffer.Len())
	return io.NopCloser(&buffer), nil
}

func (cam *Camera) encodeVideo(inputBuf *bytes.Buffer, duration int) (io.ReadCloser, error) {
	// Временный файл для вывода
	file, err := os.CreateTemp("", "*.mp4")
	if err != nil {
		return nil, fmt.Errorf("Failed to create temporary file: %w", err)
	}

	fileName := file.Name()
	file.Close()
	defer os.Remove(fileName)

	// Команда ffmpeg
	err = ffmpeg.
		Input("pipe:0", ffmpeg.KwArgs{
			"t": fmt.Sprintf("%d", duration),
			"f": "mjpeg",
		}).
		Output(fileName,
			ffmpeg.KwArgs{
				"vf":       scale,
				"c:v":      codec,
				"preset":   preset,
				"movflags": flag,
				"pix_fmt":  pix,
				"f":        fVideo,
				"metadata": metadata,
			}).
		WithInput(inputBuf).
		WithOutput(nil, os.Stderr).
		OverWriteOutput().
		Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to encode video: %w", err)
	}

	// Открываем файл как io.ReadCloser
	outFile, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open mp4 file: %w", err)
	}

	return outFile, nil
}
