package repository

import (
	"fmt"

	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/jmoiron/sqlx"
)

type ICamerasRepository interface {
	GetAllCameras() ([]models.Camera, error)
	GetCameraByName(name string) (models.Camera, error)
	GetCameraByRTSP(rtsp string) (models.Camera, error)
	AddCamera(camera models.Camera) error
	RemoveCamera(name string) (bool, error)
	UpdateCamera(camera models.Camera) (bool, error)
}

type camerasRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) ICamerasRepository {
	return &camerasRepository{
		db: db,
	}
}

func (repo *camerasRepository) GetAllCameras() ([]models.Camera, error) {
	const query = "SELECT name, rtsp FROM main.cameras;"

	var users []models.Camera

	if err := repo.db.Select(&users, query); err != nil {
		return nil, fmt.Errorf("Invalid SELECT main.cameras: %s", err)
	}

	return users, nil
}

func (repo *camerasRepository) GetCameraByName(name string) (models.Camera, error) {
	const query = "SELECT name, rtsp FROM main.cameras WHERE name = $1;"

	var user models.Camera
	if err := repo.db.Get(&user, query, name); err != nil {
		return models.Camera{}, fmt.Errorf("User does not exist in main.cameras: %w", err)
	}

	return user, nil
}

func (repo *camerasRepository) GetCameraByRTSP(rtsp string) (models.Camera, error) {
	const query = "SELECT name, rtsp FROM main.cameras WHERE rtsp = $1;"

	var user models.Camera
	if err := repo.db.Get(&user, query, rtsp); err != nil {
		return models.Camera{}, fmt.Errorf("Camera does not exist in main.cameras: %w", err)
	}

	return user, nil
}

func (repo *camerasRepository) AddCamera(user models.Camera) error {
	const query = `
INSERT INTO main.cameras (
	name,
	rtsp
) VALUES (
	:name,
	:rtsp
) ON CONFLICT (rtsp) DO NOTHING;`

	resp, err := repo.db.NamedExec(query, user)
	if err != nil {
		return fmt.Errorf("Invalid INSERT INTO main.cameras: %s", err)
	}

	count, err := resp.RowsAffected()
	if err != nil {
		return fmt.Errorf("Invalid affected camera: %s", err)
	}

	if count == 0 {
		return fmt.Errorf("Camera already exists in main.cameras")
	}

	return nil
}

func (repo *camerasRepository) RemoveCamera(name string) (bool, error) {
	const query = "DELETE FROM main.cameras WHERE name = $1"

	exec, err := repo.db.Exec(query, name)
	if err != nil {
		return false, fmt.Errorf("Invalid DELETE main.cameras: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Invalid affected camera: %s", err)
	}

	return affected == 0, nil
}

func (repo *camerasRepository) UpdateCamera(camera models.Camera) (bool, error) {
	const query = `
UPDATE main.cameras
SET name = :name,
	rtsp = :rtsp
WHERE rtsp = :rtsp;`

	exec, err := repo.db.NamedExec(query, camera)
	if err != nil {
		return false, fmt.Errorf("Invalid UPDATE main.cameras: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Invalid affected user: %s", err)
	}

	return affected == 1, nil
}
