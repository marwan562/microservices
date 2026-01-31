package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// sqlDBWrapper wraps *sql.DB to satisfy the DB interface.
type sqlDBWrapper struct {
	*sql.DB
}

func (w *sqlDBWrapper) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return w.DB.QueryRowContext(ctx, query, args...)
}

// User represents a registered user in the system.
type User struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id,omitempty"` // Primary or current organization
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never return password
	CreatedAt time.Time `json:"created_at"`
}

// Repository handles database interactions for users.
type Repository struct {
	db DB
}

// NewRepository creates a new instance of Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: &sqlDBWrapper{db}}
}

// NewTestRepository creates a repository with a custom DB interface for testing.
func NewTestRepository(db DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user into the database and returns the created user.
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at",
		email, passwordHash).Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, created_at FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
