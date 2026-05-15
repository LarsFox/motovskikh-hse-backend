package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager/auth"
)

var (
	ErrInvalidTokenClaims      = errors.New("invalid token claims")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

const (
	refreshTokenTTL = 365 * 24 * time.Hour
)

func (m *Manager) RefreshToken(ctx context.Context, token string) (*entities.TokenPair, error) {
	hash := auth.HashToken(token)
	userID, err := m.db.GetRefreshToken(ctx, hash)

	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return nil, ErrInvalidRefreshToken
	default:
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	fresh, err := auth.GeneratePair(userID, m.secretKey)
	if err != nil {
		return nil, fmt.Errorf("generate pair: %w", err)
	}

	// Помечаем старый токен использованным.
	if err := m.db.RefreshToken(ctx, hash, fresh.RefreshToken, time.Now().Add(refreshTokenTTL)); err != nil {
		return nil, fmt.Errorf("mark token as used: %w", err)
	}

	return fresh, nil
}

// ValidateAccess проверяет access токен и возвращает идентификатор пользователя.
func (m *Manager) ValidateAccess(tokenString string) (int64, error) {
	userID, err := auth.ValidateAccessToken(tokenString, m.secretKey)
	if err != nil {
		return 0, fmt.Errorf("validate access token: %w", err)
	}

	return userID, nil
}

// GenerateTokensByEmail выдаёт пару токенов по email без проверки пароля.
// Используется после верификации email.
func (m *Manager) GenerateTokensByEmail(ctx context.Context, email string) (*entities.TokenPair, error) {
	user, err := m.db.GetUserByEmail(ctx, email)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return nil, entities.ErrInvalidInput
	default:
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	tokens, err := auth.GeneratePair(user.ID, m.secretKey)
	if err != nil {
		return nil, fmt.Errorf("generate pair: %w", err)
	}
	return tokens, nil
}
