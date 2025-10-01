package entities

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name,omitempty"`
	Email        string    `json:"email" validate:"required,email"`
	PasswordHash string    `json:"password_hash,omitempty" validate:"required"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type CreateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserId    string    `json:"user_id" validate:"required"`
	TokenHash string    `json:"token_hash" validate:"required"`
	ExpiresAt time.Time `json:"expires_at,omitempty" validate:"required"`
	Revoked   bool      `json:"revoked,omitempty"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
