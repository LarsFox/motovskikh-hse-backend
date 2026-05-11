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
	// TODO: switch ErrNotFound default
	if err != nil {
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
		return 0, err // TODO: fmt.Errorf
	}

	return userID, nil
}
