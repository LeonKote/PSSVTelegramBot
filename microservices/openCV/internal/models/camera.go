package models

type Camera struct {
	Name string `db:"name" json:"name"`
	Rtsp string `db:"rtsp" json:"rtsp"`
}
