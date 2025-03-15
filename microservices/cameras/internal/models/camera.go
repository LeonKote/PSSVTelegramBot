package models

type Camera struct {
	Id   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Mac  string `db:"mac" json:"mac"`
} // @name camera
