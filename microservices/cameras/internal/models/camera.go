package models

type Camera struct {
	Name string `db:"name" json:"name"`
	Rtsp string `db:"rtsp" json:"rtsp"`
} // @name camera

type Record struct {
	ChatID     int64  `json:"chat_id"`
	NameCamera string `json:"name"`
	Duration   *int   `json:"duration"`
} // @name record

type RTSP struct {
	Rtsp string `db:"rtsp" json:"rtsp"`
} // @name rtsp


