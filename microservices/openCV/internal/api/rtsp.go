package api

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/Impisigmatus/service_core/log"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Camera struct {
	rtspUrl string
}

func NewCamera(rtspURL string) *Camera {
	_, err := ffmpeg.ProbeWithTimeout(rtspURL, 10*time.Second, ffmpeg.KwArgs{
		"fflags":         "+genpts",
		"rtsp_transport": "tcp",
	})
	if err != nil {
		log.Errorf("Failed to probe RTSP stream: %v", err)
		return nil
	}

	return &Camera{
		rtspUrl: rtspURL,
	}
}

func (cam *Camera) CapturePhoto() ([]byte, error) {
	begin := time.Now()
	log.Debugf("Start capture photo at: %s", begin)
	var buffer bytes.Buffer
	err := ffmpeg.Input(cam.rtspUrl, ffmpeg.KwArgs{
		"fflags":   "+genpts",
		"loglevel": "error",
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

	log.Infof("End capture photo. Duration: %s, len: %d", time.Since(begin), buffer.Len())

	return buffer.Bytes(), nil
}
