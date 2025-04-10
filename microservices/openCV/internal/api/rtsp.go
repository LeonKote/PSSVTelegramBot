package api

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Camera struct {
	rtspUrl string
}

func NewCamera(log zerolog.Logger, rtspURL string, auth string) *Camera {
	_, err := ffmpeg.ProbeWithTimeout(rtspURL, 10*time.Second, ffmpeg.KwArgs{
		"fflags":  "+genpts",
		"headers": auth,
	})
	if err != nil {
		log.Panic().Msgf("Failed to probe RTSP stream: %s", err)
	}

	return &Camera{
		rtspUrl: rtspURL,
	}
}

func (cam *Camera) CapturePhoto(log zerolog.Logger, auth string) ([]byte, error) {
	begin := time.Now()
	log.Debug().Msgf("Start capture photo at: %s", begin)
	var buffer bytes.Buffer
	err := ffmpeg.Input(cam.rtspUrl, ffmpeg.KwArgs{
		"fflags":   "+genpts",
		"loglevel": "error",
		"headers":  auth,
	}).
		Output("pipe:1",
			ffmpeg.KwArgs{
				"f":        "mjpeg",
				"frames:v": "1",
				"q:v":      "2",
			}).
		WithOutput(&buffer).
		WithErrorOutput(os.Stderr).
		Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to encode photo: %w", err)
	}

	log.Info().Msgf("End capture photo. Duration: %s, len: %d", time.Since(begin), buffer.Len())

	return buffer.Bytes(), nil
}
