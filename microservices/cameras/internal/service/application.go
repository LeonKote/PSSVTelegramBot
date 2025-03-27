package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	log "github.com/Impisigmatus/service_core/log"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/repository"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Application struct {
	ctx         context.Context
	repo        repository.ICamerasRepository
	cams        map[string]*api.Camera
	minioClient *minio.Client
	bucketName  string
	apiFiles    *api.FilesAPI
	cfg         config.Config
}

func MakeApplication(ctx context.Context, db *sqlx.DB, cfg config.Config) *Application {
	cams := make(map[string]*api.Camera)
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Panicf("Invalid connect to minio: %s", err)
	}

	camerasRepository := repository.NewRepository(db)
	cameras, err := camerasRepository.GetAllCameras()
	if err != nil {
		log.Panicf("Invalid get cameras: %s", err)
	}

	for _, camera := range cameras {
		cams[camera.Name] = api.NewCamera(camera.Rtsp)
	}

	return &Application{
		ctx:         ctx,
		repo:        camerasRepository,
		cams:        cams,
		minioClient: minioClient,
		bucketName:  cfg.BucketName,
		apiFiles:    api.NewFilesApi(cfg),
		cfg:         cfg,
	}
}

func (app *Application) Record(ctx context.Context, record models.Record, isVideo bool, reqId string) (string, error) {
	cam, err := app.repo.GetCameraByName(record.NameCamera)
	if err != nil {
		return "", fmt.Errorf("Invalid camera name: %s", err)
	}

	var fileName string
	var fileType string
	if isVideo {
		fileName = fmt.Sprintf(fileNameVideo, record.ChatID, reqId)
		fileType = fileTypeVideo
	} else {
		fileName = fmt.Sprintf(fileNameImage, record.ChatID, reqId)
		fileType = fileTypeImage
	}

	file := models.File{
		ChatID:     record.ChatID,
		CameraName: record.NameCamera,
		Uuid:       reqId,
		FilePath:   fileName,
		FileSize:   nilFileSize,
		Status:     statusPending,
		FileType:   fileType,
	}

	log.Debugf("Len: %d.", file.FileSize)
	ok, err := app.apiFiles.AddFile(file)
	if err != nil {
		return fileName, fmt.Errorf("Invalid load file: %s", err)
	}

	if !ok {
		return fileName, fmt.Errorf("File already exist: %s", err)
	}

	streamUrl := fmt.Sprintf("%s/%s", app.cfg.StreamUrl, cam.Name)

	var buff io.ReadCloser
	if isVideo {
		buff, err = app.cams[cam.Name].RecordVideo(ctx, *record.Duration, streamUrl)
		if err != nil {
			return fileName, fmt.Errorf("Invalid record video: %s", err)
		}
	} else {
		buff, err = app.cams[cam.Name].CapturePhoto(streamUrl)
		if err != nil {
			return fileName, fmt.Errorf("Invalid capture photo: %s", err)
		}
	}

	defer buff.Close()

	data, err := io.ReadAll(buff)
	if err != nil {
		return fileName, fmt.Errorf("Invalid read body: %s", err)
	}

	lenData := len(data)

	if lenData == 0 {
		return fileName, fmt.Errorf("Invalid make file: %s", err)
	}

	log.Debugf("Len: %d. Data: %s", lenData, data)
	ok, err = app.AttemptLoad(data, fileName, lenData, ctx, app.bucketName)
	if err != nil {
		return fileName, fmt.Errorf("Can not load file: %w", err)
	}
	if !ok {
		return fileName, fmt.Errorf("File already exist: %w", err)
	}

	if err := app.ChangeStatus(fileName, lenData, statusReady); err != nil {
		return fileName, fmt.Errorf("Invalid load file: %w", err)
	}

	return fileName, nil
}

func (app *Application) ChangeStatus(fileName string, fileSize int, status string) error {
	if err := app.apiFiles.ChangeStatus(fileName, fileSize, status); err != nil {
		return fmt.Errorf("Invalid add queue: %w", err)
	}

	return nil
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
