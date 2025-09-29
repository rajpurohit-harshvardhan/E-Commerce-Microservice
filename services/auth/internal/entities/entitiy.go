package entities

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name,omitempty"`
	Email        string    `json:"email" validate:"required,email"`
	PasswordHash string    `json:"password_hash,omitempty" validate:"required"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
