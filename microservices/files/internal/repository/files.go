package repository

import (
	"fmt"

	_ "github.com/Impisigmatus/service_core/postgres"
	"github.com/LeonKote/PSSVTelegramBot/microservices/files/internal/models"
	"github.com/jmoiron/sqlx"
)

type IFileRepo interface {
	GetAllFiles() ([]models.File, error)
	GetFileByUuid(uuid string) (models.File, error)
	AddFile(file models.File) error
	RemoveFile(uuid string) error
	UpdateFileData(filePath string, status string, fileSize int) error
	GetReadyFile() ([]models.File, error)
}

type fileRepo struct {
	db *sqlx.DB
}

func NewFileRepo(db *sqlx.DB) IFileRepo {
	return &fileRepo{
		db: db,
	}
}

func (f *fileRepo) GetAllFiles() ([]models.File, error) {
	const query = `
SELECT chat_id, camera_name, uuid, file_path, file_size, file_type, status, captured_at
FROM main.file_metadata;`

	var files []models.File

	if err := f.db.Select(&files, query); err != nil {
		return nil, fmt.Errorf("Invalid SELECT main.file_metadata: %s", err)
	}

	return files, nil
}

func (f *fileRepo) GetFileByUuid(uuid string) (models.File, error) {
	const query = `
SELECT chat_id, camera_name, uuid, file_path, file_size, file_type, status, captured_at
FROM main.file_metadata
WHERE uuid = $1;`

	var file models.File
	if err := f.db.Get(&file, query, uuid); err != nil {
		return models.File{}, fmt.Errorf("File does not exist in main.file_metadata: %w", err)
	}

	return file, nil
}

func (f *fileRepo) AddFile(file models.File) error {
	const query = `
INSERT INTO main.file_metadata (
	chat_id,
	camera_name,
	uuid,
	file_path,
	file_size,
	file_type
) VALUES (
	:chat_id,
	:camera_name,
	:uuid,
	:file_path,
	:file_size,
	:file_type
) ON CONFLICT (uuid) DO NOTHING;`

	if _, err := f.db.NamedExec(query, file); err != nil {
		return fmt.Errorf("Invalid INSERT INTO main.file_metadata: %s", err)
	}

	return nil
}

func (f *fileRepo) RemoveFile(uuid string) error {
	const query = "DELETE FROM main.file_metadata WHERE uuid = $1"

	exec, err := f.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("Invalid DELETE main.file_metadata: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return fmt.Errorf("Invalid affected file: %s", err)
	}
	if affected == 0 {
		return NoAffectedError
	}

	return nil
}

func (f *fileRepo) UpdateFileData(filePath string, status string, fileSize int) error {
	const query = "UPDATE main.file_metadata SET status = $1, file_size = $2 WHERE file_path = $3;"

	exec, err := f.db.Exec(query, status, fileSize, filePath)
	if err != nil {
		return fmt.Errorf("Invalid UPDATE main.file_metadata: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return fmt.Errorf("Invalid affected user: %s", err)
	}
	if affected == 0 {
		return NoAffectedError
	}

	return nil
}

func (f *fileRepo) GetReadyFile() ([]models.File, error) {
	const query = "SELECT chat_id, camera_name, uuid, file_path, file_size, file_type, status FROM main.file_metadata WHERE status = 'ready';"
	var files []models.File
	if err := f.db.Select(&files, query); err != nil {
		return nil, fmt.Errorf("Invalid SELECT main.file_metadata: %s", err)
	}

	return files, nil
}
