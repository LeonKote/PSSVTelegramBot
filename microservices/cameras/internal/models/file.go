package models

import (
	"time"
)

type File struct {
	ChatID     int64      `db:"chat_id" json:"chat_id"`
	CameraName string     `db:"camera_name" json:"camera_name"`
	Uuid       string     `db:"uuid" json:"uuid"`
	FilePath   string     `db:"file_path" json:"file_path"`
	FileSize   int        `db:"file_size" json:"file_size"`
	FileType   string     `db:"file_type" json:"file_type"`
	Status     string     `db:"status" json:"status"`
	CapturedAt *time.Time `db:"captured_at" json:"captured_at"`
} // @name file

type Uuid struct {
	Uuid string `json:"uuid"`
}
