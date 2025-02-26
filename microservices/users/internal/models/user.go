package models

type User struct {
	Chat_ID  int64  `db:"chat_id" json:"chat_id"`
	Username string `db:"username" json:"username"`
	Name     string `db:"name" json:"name"`
	Is_Admin bool   `db:"is_admin" json:"is_admin"`
	Status   string `db:"status" json:"status"`
} // @name user
