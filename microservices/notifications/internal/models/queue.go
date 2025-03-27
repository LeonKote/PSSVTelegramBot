package models

type Queue struct {
	ChatID   int64  `db:"chat_id" json:"chat_id"`
	FilePath string `db:"file_path" json:"file_path"`
	Status   string `db:"status" json:"status"`
} // @name queue
