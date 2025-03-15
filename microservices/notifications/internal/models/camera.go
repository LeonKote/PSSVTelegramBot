package models

type Camera struct {
	Id   int64  `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	Mac  string `json:"mac" validate:"required"`
}
