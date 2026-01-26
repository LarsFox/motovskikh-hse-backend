package repository

import (
	"context"
	"database/sql"

	"coursework/internal/models"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) Create(
	ctx context.Context,
	email string,
) (*models.User, error) {

	result, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (email) VALUES (?)",
		email,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:            id,
		Email:         email,
		EmailVerified: false,
	}, nil
}
