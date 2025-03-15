package database

import (
	"fmt"

	_ "github.com/Impisigmatus/service_core/postgres"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/models"
	"github.com/jmoiron/sqlx"
)

type Database struct {
	db *sqlx.DB
}

func NewDatabase(db *sqlx.DB) *Database {
	return &Database{
		db: db,
	}
}

func (pg *Database) GetAllCameras() ([]models.Camera, error) {
	const query = "SELECT id, name, mac FROM main.cameras;"

	var users []models.Camera

	if err := pg.db.Select(&users, query); err != nil {
		return nil, fmt.Errorf("Invalid SELECT main.cameras: %s", err)
	}

	return users, nil
}

func (pg *Database) GetCameraByID(id int) (models.Camera, error) {
	const query = "SELECT id, name, mac FROM main.cameras WHERE id = $1;"

	var user models.Camera
	if err := pg.db.Get(&user, query, id); err != nil {
		return models.Camera{}, fmt.Errorf("User does not exist in main.cameras: %w", err)
	}

	return user, nil
}

func (pg *Database) AddCamera(user models.Camera) error {
	const query = `
INSERT INTO main.cameras (
	name,
	mac
) VALUES (
	:name,
	:mac
) ON CONFLICT (mac) DO NOTHING;`

	if _, err := pg.db.NamedExec(query, user); err != nil {
		return fmt.Errorf("Invalid INSERT INTO main.cameras: %s", err)
	}

	return nil
}

func (pg *Database) RemoveCamera(id int) (bool, error) {
	const query = "DELETE FROM main.cameras WHERE id = $1"

	exec, err := pg.db.Exec(query, id)
	if err != nil {
		return false, fmt.Errorf("Invalid DELETE main.cameras: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Invalid affected camera: %s", err)
	}

	return affected == 0, nil
}
