package app

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/Impisigmatus/service_core/log"
	"github.com/LeonKote/PSSVTelegramBot/microservices/openCV/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/openCV/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gocv.io/x/gocv"
)

type Application struct {
	camera      *api.Camera
	cfg         config.Config
	minioClient *minio.Client
	notifyApi   *api.NotifyAPI
}

func MakeApplication(cfg config.Config) *Application {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Panicf("Invalid connect to minio: %s", err)
	}

	return &Application{
		camera:      api.NewCamera(cfg.StreamUrl),
		cfg:         cfg,
		minioClient: minioClient,
		notifyApi:   api.NewNotifyAPI(cfg.NotificationsApi),
	}
}

func (app *Application) CheckPhoto(ctx context.Context) {
	prevFrame := gocv.NewMat()

	for {
		newFrame, err := app.processing(ctx, prevFrame)
		if err != nil {
			log.Errorf("Failed to process photo: %s", err)
		}

		prevFrame = newFrame
	}
}

func (app *Application) processing(ctx context.Context, prevFrame gocv.Mat) (gocv.Mat, error) {
	data, err := app.camera.CapturePhoto()
	if data == nil {
		return gocv.Mat{}, fmt.Errorf("Failed to capture photo: %w", err)
	}

	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil || img.Empty() {
		return gocv.Mat{}, fmt.Errorf("Failed to decode photo: %w", err)
	}
	defer img.Close()

	// Если это первый кадр — возвращаем как новый prev
	if prevFrame.Empty() {
		return img.Clone(), nil
	}
	defer prevFrame.Close()

	// Переводим оба кадра в grayscale
	prevGray := gocv.NewMat()
	currGray := gocv.NewMat()
	defer prevGray.Close()
	defer currGray.Close()

	gocv.CvtColor(prevFrame, &prevGray, gocv.ColorBGRToGray)
	gocv.CvtColor(img, &currGray, gocv.ColorBGRToGray)

	// Вычисляем разницу
	diff := gocv.NewMat()
	defer diff.Close()

	gocv.AbsDiff(prevGray, currGray, &diff)

	// Считаем количество отличий
	nonZero := gocv.CountNonZero(diff)
	log.Infof("Count non zero: %d", nonZero)

	// Настраиваем порог чувствительности (увеличен для меньшей чувствительности)
	if nonZero > 800000 {
		filePath := fmt.Sprintf("alert/%d.png", time.Now().Unix())
		ok, err := app.AttemptLoad(data,
			filePath,
			len(data),
			ctx,
			app.cfg.BucketName,
		)
		if err != nil || !ok {
			return gocv.Mat{}, fmt.Errorf("Failed to attempt load: %w", err)
		}

		ok, err = app.notifyApi.SendAlert(filePath)
		if err != nil || !ok {
			return gocv.Mat{}, fmt.Errorf("Failed to send alert: %w", err)
		}
	}

	return img.Clone(), nil
}

func (app *Application) AttemptLoad(data []byte, fileName string, len int, ctx context.Context, bucketName string) (bool, error) {
	maxRetry := 3
	for i := 1; i <= maxRetry; i++ {
		tmpData := bytes.NewBuffer(data)

		var writer bytes.Buffer
		tee := io.TeeReader(tmpData, &writer)

		log.Debugf("Start load file: %s, bucketName: %s, fileName: %s, len: %d", time.Now(), bucketName, fileName, len)
		info, err := app.minioClient.PutObject(ctx,
			bucketName,
			fileName,
			tee,
			int64(len),
			minio.PutObjectOptions{
				ContentType: "application/octet-stream",
			},
		)
		if err != nil {
			return false, fmt.Errorf("Invalid load file: %s", err)
		}

		if info.Size == 0 {
			return false, ErrSizeZero
		}

		hash := md5.Sum(data)
		hashString := hex.EncodeToString(hash[:])
		if hashString != info.ETag {
			if i == 3 {
				return false, fmt.Errorf("Different hash")
			}

			continue
		}

		break
	}

	return true, nil
}
