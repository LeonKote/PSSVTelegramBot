package models

type Camera struct {
	Name string `json:"name" validate:"required"`
	Rtsp string `json:"rtsp" validate:"required"`
}

type Record struct {
	ChatID     int64  `json:"chat_id"`
	NameCamera string `json:"name"`
	Duration   *int   `json:"duration"`
}
