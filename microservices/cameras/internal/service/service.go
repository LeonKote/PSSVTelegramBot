package service

import (
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/database"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db database.Database
}

func NewService(db *sqlx.DB) *Service {
	return &Service{
		db: *database.NewDatabase(db),
	}
}

func (srv *Service) GetAllCameras() ([]models.Camera, error) {
	return srv.db.GetAllCameras()
}

func (srv *Service) GetCameraByID(id int) (models.Camera, error) {
	return srv.db.GetCameraByID(id)
}

func (srv *Service) AddCamera(user models.Camera) error {
	return srv.db.AddCamera(user)
}

func (srv *Service) RemoveCamera(id int) (bool, error) {
	return srv.db.RemoveCamera(id)
}
