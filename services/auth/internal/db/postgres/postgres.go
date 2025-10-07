package postgres

import (
	"auth/internal/entities"
	"common/utils/migrate"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"strings"

	"auth/internal/config"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

var embedMigrations embed.FS

func New(cfg *config.Config) (*Postgres, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Cockroach.User,
		cfg.Cockroach.Password,
		cfg.Cockroach.Host,
		cfg.Cockroach.Port,
		cfg.Cockroach.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure the connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Running migrations
	if err := migrate.Run(db, embedMigrations, "."); err != nil {

		if strings.Contains(err.Error(), "no migration files found") {
			slog.Warn("migrations: none found; continuing")
		} else {
			_ = db.Close()
			return nil, fmt.Errorf("migrations failed: %w", err)
		}
	}

	return &Postgres{
		Db: db,
	}, nil
}

func (p *Postgres) CreateUser(name string, email string, passwordHash string) (string, error) {
	slog.Debug("CreateUser DB :: ", slog.String("name", name), slog.String("email", email))
	var id string

	err := p.Db.QueryRow(
		`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		name, email, passwordHash).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *Postgres) DeleteUserById(id string) (bool, error) {
	slog.Debug("DeleteUserById DB :: ", slog.String("id", id))
	result, err := p.Db.Exec(`DELETE FROM users where id=$1`, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to delete User: ", err)
		return false, nil
	}

	return true, nil
}

func (p *Postgres) GetUserById(id string) (entities.User, error) {
	slog.Debug("GetUserById DB :: ", slog.String("id", id))
	var user entities.User
	err := p.Db.QueryRow("SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id=$1",
		id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (p *Postgres) UpdateUserById(id string, detailsToUpdate map[string]interface{}) (bool, error) {
	slog.Debug("UpdateUserById DB :: ", slog.String("id", id))

	if len(detailsToUpdate) == 0 {
		return false, errors.New("no fields to update")
	}

	validColumns := map[string]bool{
		"name":          true,
		"email":         true,
		"password_hash": true,
	}

	var fields []string
	var values []interface{}
	placeholderCounter := 1

	for key, value := range detailsToUpdate {
		if !validColumns[key] {
			return false, fmt.Errorf("invalid field: %s", key)
		}
		fields = append(fields, fmt.Sprintf("%s = $%d", key, placeholderCounter))
		values = append(values, value)
		placeholderCounter++
	}

	fields = append(fields, fmt.Sprintf("updated_at = $%d", placeholderCounter))
	values = append(values, time.Now())
	placeholderCounter++

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(fields, ", "), placeholderCounter)
	values = append(values, id)

	result, err := p.Db.Exec(query, values...)
	if err != nil {
		return false, fmt.Errorf("failed to update auth: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to update user: ", err)
		return false, fmt.Errorf("user not found or no changes made: %w", err)
	}

	return true, nil
}

func (p *Postgres) GetUserByEmail(email string) (entities.User, error) {
	slog.Debug("GetUserByEmail DB :: ", slog.String("email", email))

	var user entities.User
	err := p.Db.QueryRow("SELECT id, name, email, password_hash FROM users WHERE email=$1",
		email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		return user, err
	}
	return user, nil
}

// RefreshTokens Table
func (p *Postgres) CreateRefreshToken(userId string, tokenHash string, expiresAt time.Time, issuedAt time.Time, revoked bool) (string, error) {
	slog.Debug("CreateRefreshToken DB :: ", slog.String("userId", userId))

	var id string

	err := p.Db.QueryRow(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, issued_at, revoked) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userId, tokenHash, expiresAt, issuedAt, revoked).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *Postgres) GetTokenByUserId(userId string) (entities.RefreshToken, error) {
	slog.Debug("GetTokenByUserId DB :: ", slog.String("userId", userId))

	var token entities.RefreshToken
	err := p.Db.QueryRow("SELECT id, user_id, token_hash, expires_at, issued_at, revoked, created_at FROM refresh_tokens WHERE user_id=$1",
		userId).Scan(&token.ID, &token.UserId, &token.TokenHash, &token.ExpiresAt, &token.IssuedAt, &token.Revoked, &token.CreatedAt)
	if err != nil {
		return token, err
	}
	return token, nil
}

func (p *Postgres) DeleteTokenById(id string) (bool, error) {
	slog.Debug("DeleteTokenById DB :: ", slog.String("id", id))

	result, err := p.Db.Exec(`DELETE FROM refresh_tokens where id=$1`, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to delete token: ", err)
		return false, nil
	}

	return true, nil
}

func (p *Postgres) DeleteTokenByUserId(userId string) (bool, error) {
	slog.Debug("DeleteTokenByUserId DB :: ", slog.String("userId", userId))

	result, err := p.Db.Exec(`DELETE FROM refresh_tokens where user_id=$1`, userId)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to delete token: ", err)
		return false, nil
	}

	return true, nil
}

func (p *Postgres) GetTokenByHash(hash string) (entities.RefreshToken, error) {
	slog.Debug("GetTokenByHash DB :: ", slog.String("hash", hash))

	var token entities.RefreshToken
	err := p.Db.QueryRow("SELECT id, user_id, token_hash, expires_at, issued_at, revoked, created_at FROM refresh_tokens WHERE token_hash=$1",
		hash).Scan(&token.ID, &token.UserId, &token.TokenHash, &token.ExpiresAt, &token.IssuedAt, &token.Revoked, &token.CreatedAt)
	if err != nil {
		return token, err
	}
	return token, nil
}

func (p *Postgres) DeleteTokenByHash(hash string) (bool, error) {
	slog.Debug("DeleteTokenByHash DB :: ", slog.String("hash", hash))

	result, err := p.Db.Exec(`DELETE FROM refresh_tokens where token_hash=$1`, hash)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Warn("failed to delete token: ", err)
		return false, nil
	}

	return true, nil
}
