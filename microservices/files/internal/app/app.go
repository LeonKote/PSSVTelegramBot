package app

import (
	"context"
	"time"

	"github.com/Impisigmatus/service_core/log"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/models"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/repository"
	"github.com/jmoiron/sqlx"
)

type Application struct {
	fileRepo repository.IFileRepo
	notify   api.INotifyAPI
}

func NewApp(cfg config.Config, db *sqlx.DB) *Application {
	return &Application{
		fileRepo: repository.NewFileRepo(db),
		notify:   api.NewNotifyApi(cfg),
	}
}

func (app *Application) CheckTable(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("CheckTable остановлен")
			return

		case <-ticker.C:
			files, err := app.GetAllFilesReady()
			if err != nil {
				log.Error("Can not get all files with arg 'ready': %s", err)
				continue
			}

			for _, file := range files {
				if err := app.fileRepo.UpdateFileData(file.FilePath, "processing", file.FileSize); err != nil {
					log.Errorf("ошибка обновления файла: %s", err)
				}

				go func(f models.File) {
					notify := models.Notify{
						ChatID:   f.ChatID,
						FilePath: f.FilePath,
						Format:   f.FileType,
					}

					if err := app.notify.Notify(notify); err != nil {
						log.Errorf("ошибка уведомления: %s", err)
					}

					if err := app.fileRepo.UpdateFileData(f.FilePath, "success", f.FileSize); err != nil {
						log.Errorf("ошибка обновления файла: %s", err)
					}
				}(file)
			}
		}
	}
}

func (app *Application) AddFile(file models.File) error {
	if err := app.fileRepo.AddFile(file); err != nil {
		return err
	}

	return nil
}

func (app *Application) RemoveFile(filePath string) error {
	if err := app.fileRepo.RemoveFile(filePath); err != nil {
		return err
	}

	return nil
}

func (app *Application) GetByUuid(uuid string) (models.File, error) {
	file, err := app.fileRepo.GetFileByUuid(uuid)
	if err != nil {
		return models.File{}, err
	}

	return file, nil
}

func (app *Application) GetAllFiles() ([]models.File, error) {
	files, err := app.fileRepo.GetAllFiles()
	if err != nil {
		return []models.File{}, err
	}

	return files, nil
}

func (app *Application) UpdateFile(filePath string, status string, size int) error {
	if err := app.fileRepo.UpdateFileData(filePath, status, size); err != nil {
		return err
	}

	return nil
}

func (app *Application) GetAllFilesReady() ([]models.File, error) {
	files, err := app.fileRepo.GetReadyFile()
	if err != nil {
		return []models.File{}, err
	}

	return files, nil
}
