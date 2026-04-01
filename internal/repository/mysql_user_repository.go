package repository

import (
	"coursework/internal/models"
	"database/sql"
	"fmt"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, login, password_hash, email_verified)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, user.Email, user.Login, user.PasswordHash, user.EmailVerified)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = int(id)
	return nil
}

func (r *MySQLUserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, login, password_hash, email_verified, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Login, &user.PasswordHash,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *MySQLUserRepository) GetByLogin(login string) (*models.User, error) {
	query := `
		SELECT id, email, login, password_hash, email_verified, created_at, updated_at
		FROM users
		WHERE login = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, login).Scan(
		&user.ID, &user.Email, &user.Login, &user.PasswordHash,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *MySQLUserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, email, login, password_hash, email_verified, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Login, &user.PasswordHash,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *MySQLUserRepository) UpdateEmailVerified(userID int, verified bool) error {
	query := `UPDATE users SET email_verified = ? WHERE id = ?`
	_, err := r.db.Exec(query, verified, userID)
	if err != nil {
		return fmt.Errorf("failed to update email_verified: %w", err)
	}
	return nil
}

func (r *MySQLUserRepository) UpdatePassword(userID int, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE id = ?`
	_, err := r.db.Exec(query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}
