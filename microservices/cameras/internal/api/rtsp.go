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
	// Контекст с таймаутом на всю запись
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(duration)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, sourceStreamURL, nil)
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

	log.Info("Копирование завершено")
	buff, err := cam.encodeVideo(&buffer, duration)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (cam *Camera) CapturePhoto(sourceStreamURL string) (io.ReadCloser, error) {
	log.Debugf("Start capture photo at: %s", time.Now())
	var buffer bytes.Buffer
	log.Debugf("Url: %s, frame: %d, quality: %d, format: %s", sourceStreamURL, frame, quality, fImg)
	err := ffmpeg.Input(sourceStreamURL, ffmpeg.KwArgs{
		"fflags": "+genpts", // добавим безопасный флаг
		"f":      "mjpeg",   // явно указываем, что читаем MJPEG
	}).
		Output("pipe:1",
			ffmpeg.KwArgs{
				"f":        "image2", // формат — один файл (image2)
				"c:v":      "mjpeg",
				"frames:v": "1",
			}).
		WithOutput(&buffer).
		WithErrorOutput(os.Stderr).
		Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to encode photo: %w", err)
	}

	log.Infof("End capture photo at: %s, len: %d", time.Now(), buffer.Len())
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
