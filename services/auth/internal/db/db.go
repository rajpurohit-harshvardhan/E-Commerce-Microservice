package db

import "auth/internal/entities"

type Db interface {
	CreateUser(name string, email string, passwordHash string) (string, error)
	DeleteUserById(id string) (bool, error)
	GetUserById(id string) (entities.User, error)
	UpdateUserById(id string, detailsToUpdate map[string]interface{}) (bool, error)
}
