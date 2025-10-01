package db

import (
	"auth/internal/entities"
	"time"
)

type Db interface {
	CreateUser(name string, email string, passwordHash string) (string, error)
	DeleteUserById(id string) (bool, error)
	GetUserById(id string) (entities.User, error)
	UpdateUserById(id string, detailsToUpdate map[string]interface{}) (bool, error)
	GetUserByEmail(email string) (entities.User, error)

	//	RefreshTokens table
	CreateRefreshToken(userId string, tokenHash string, expiresAt time.Time, issuedAt time.Time, revoked bool) (string, error)
	GetTokenByUserId(userId string) (entities.RefreshToken, error)
	DeleteTokenById(id string) (bool, error)
	DeleteTokenByUserId(userId string) (bool, error)
	GetTokenByHash(hash string) (entities.RefreshToken, error)
	DeleteTokenByHash(hash string) (bool, error)
}
