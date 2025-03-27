package repository

import (
	"fmt"

	_ "github.com/Impisigmatus/service_core/postgres"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/internal/models"
	"github.com/jmoiron/sqlx"
)

type IUsersRepository interface {
	GetAllUsers() ([]models.User, error)
	GetUserByChatID(chatId int64) (models.User, error)
	AddUser(user models.User) error
	RemoveUser(chatId int64) (bool, error)
	GetAdminUser() (models.User, error)
	UpdateUser(user models.User) (bool, error)
}

type usersRepository struct {
	db *sqlx.DB
}

func NewUsersRepository(db *sqlx.DB) IUsersRepository {
	return &usersRepository{
		db: db,
	}
}

func (r *usersRepository) GetAllUsers() ([]models.User, error) {
	const query = "SELECT chat_id, username, name, is_admin, status FROM main.users;"

	var users []models.User

	if err := r.db.Select(&users, query); err != nil {
		return nil, fmt.Errorf("Invalid SELECT main.users: %s", err)
	}

	return users, nil
}

func (r *usersRepository) GetUserByChatID(chatId int64) (models.User, error) {
	const query = "SELECT chat_id, username, name, is_admin, status FROM main.users WHERE chat_id = $1;"

	var user models.User
	if err := r.db.Get(&user, query, chatId); err != nil {
		return models.User{}, fmt.Errorf("User does not exist in main.users: %w", err)
	}

	return user, nil
}

func (r *usersRepository) AddUser(user models.User) error {
	const query = `
INSERT INTO main.users (
	chat_id,
	username,
	name,
	is_admin,
	status
) VALUES (
	:chat_id,
	:username,
	:name,
	:is_admin,
	:status
) ON CONFLICT (chat_id) DO NOTHING;`

	if _, err := r.db.NamedExec(query, user); err != nil {
		return fmt.Errorf("Invalid INSERT INTO main.users: %s", err)
	}

	return nil
}

func (r *usersRepository) RemoveUser(chatId int64) (bool, error) {
	const query = "DELETE FROM main.users WHERE chat_id = $1"

	exec, err := r.db.Exec(query, chatId)
	if err != nil {
		return false, fmt.Errorf("Invalid DELETE main.users: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Invalid affected user: %s", err)
	}

	return affected == 0, nil
}

func (r *usersRepository) GetAdminUser() (models.User, error) {
	const query = "SELECT chat_id, username, name, is_admin, status FROM main.users WHERE is_admin = true;"

	var user models.User
	if err := r.db.Get(&user, query); err != nil {
		return models.User{}, fmt.Errorf("User does not exist in main.users: %w", err)
	}

	return user, nil
}

func (r *usersRepository) UpdateUser(user models.User) (bool, error) {
	const query = `
UPDATE main.users
SET username = :username,
	name = :name,
	status = :status
WHERE chat_id = :chat_id;`

	exec, err := r.db.NamedExec(query, user)
	if err != nil {
		return false, fmt.Errorf("Invalid UPDATE main.users: %s", err)
	}

	affected, err := exec.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Invalid affected user: %s", err)
	}

	return affected == 1, nil
}
