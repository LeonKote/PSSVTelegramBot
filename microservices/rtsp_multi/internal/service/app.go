package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Application struct {
	reader      *io.PipeReader
	writer      *io.PipeWriter
	rtspUrl     string
	mu          sync.RWMutex
	subscribers []chan []byte
}

func NewApp(rtspUrl string) *Application {
	reader, writer := io.Pipe()
	return &Application{
		reader:  reader,
		writer:  writer,
		rtspUrl: rtspUrl,
		mu:      sync.RWMutex{},
	}
}

// Захват потока через ffmpeg с возможностью остановки
func (app *Application) Run(log zerolog.Logger, ctx context.Context) error {
	log.Info().Msg("Starting FFmpeg stream via ffmpeg-go")

	a, err := ffmpeg.ProbeWithTimeout(app.rtspUrl, 10*time.Second, ffmpeg.KwArgs{
		"fflags":         "+genpts",
		"rtsp_transport": "tcp",
	})
	if err != nil {
		log.Error().Msgf("Failed to probe RTSP stream: %v", err)
		return err
	}

	log.Info().Msgf("FFmpeg stream info: %s", a)

	// Закрываем writer по завершении контекста
	go func() {
		<-ctx.Done()
		log.Info().Msg("Context canceled, closing ffmpeg writer")
		app.writer.Close()
	}()

	// Подключаем stderr для логов ffmpeg

	stderr := &bytes.Buffer{}
	cmd := ffmpeg.Input(app.rtspUrl, ffmpeg.KwArgs{
		"fflags":         "+genpts",
		"rtsp_transport": "tcp",
	}).
		Output("pipe:", ffmpeg.KwArgs{
			"f":   "mjpeg",
			"q:v": "30",
		}).
		WithErrorOutput(stderr).
		WithOutput(app.writer)

	log.Info().Msgf("FFmpeg command: %s", cmd.String())
	log.Info().Msgf("FFmpeg stderr: %s", stderr.String())

	log.Info().Msgf("Running ffmpeg with URL: %s", app.rtspUrl)

	err = cmd.Run()

	// Логируем stderr
	if stderr.Len() > 0 {
		log.Error().Msgf("FFmpeg stderr: %s", stderr.String())
	}

	if err != nil {
		log.Error().Msgf("Failed to run ffmpeg-go: %v", err)
		return err
	}
	log.Info().Msg("FFmpeg finished")

	return nil
}

// Рассылает видеопоток всем подписчикам
func (app *Application) DistributeStream(log zerolog.Logger, ctx context.Context) error {
	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			log.Error().Msg("DistributeStream stopped")
			return ctx.Err()
		default:
			n, err := app.reader.Read(buf)
			if err != nil {
				log.Error().Msgf("Pipe read error: %s", err)
				return err
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			app.mu.RLock()
			for _, ch := range app.subscribers {
				select {
				case ch <- chunk:
				default:
					log.Info().Msg("Dropping frame for slow client")
				}
			}
			app.mu.RUnlock()
		}
	}
}

// Отдаёт поток подключённому клиенту
func (app *Application) StreamHandler(log zerolog.Logger, w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("New client connected")
	w.Header().Set("Content-Type", "video/mp2t")

	clientChan := make(chan []byte, 200)

	app.mu.Lock()
	app.subscribers = append(app.subscribers, clientChan)
	log.Info().Msgf("New subscriber added: %d", len(app.subscribers))

	app.mu.Unlock()

	defer func() {
		app.mu.Lock()
		for i, ch := range app.subscribers {
			log.Info().Msgf("Removing client %d", i)
			if ch == clientChan {
				app.subscribers = append(app.subscribers[:i], app.subscribers[i+1:]...)
				break
			}
		}
		app.mu.Unlock()
		close(clientChan)
		log.Info().Msg("Client disconnected")
	}()

	// Отслеживание отключения клиента
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-clientChan:
			if !ok {
				return
			}
			_, err := w.Write(chunk)
			if err != nil {
				log.Error().Msgf("Write error: %s", err)
				return
			}

			flusher := w.(http.Flusher)
			flusher.Flush()
		}
	}
}
