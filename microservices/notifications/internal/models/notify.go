package models

type Notify struct {
	ChatID   int64  `json:"chat_id"`
	FilePath string `json:"file_path"`
	Format   string `json:"format"`
}
