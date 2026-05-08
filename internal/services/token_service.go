package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
	"github.com/LarsFox/motovskikh-hse-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

// Сервис для работы с JWT токенами
type TokenService struct {
	secretKey        []byte
	refreshTokenRepo repository.RefreshTokenRepository
}

func NewTokenService(secretKey string, refreshTokenRepo repository.RefreshTokenRepository) *TokenService {
	return &TokenService{
		secretKey:        []byte(secretKey),
		refreshTokenRepo: refreshTokenRepo,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *TokenService) GeneratePair(userID uint) (*TokenPair, error) {
	accessToken, err := s.generateAccess(userID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateRefresh(userID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccess проверяет access токен и возвращает ID пользователя
func (s *TokenService) ValidateAccess(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		// Перекус таксиста чеееек
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secretKey, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	return c.UserID, nil
}

func (s *TokenService) Refresh(refreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	stored, err := s.refreshTokenRepo.GetValid(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Помечаем старый токен использованным
	if err := s.refreshTokenRepo.MarkAsUsed(stored.ID); err != nil {
		return nil, fmt.Errorf("mark token as used: %w", err)
	}

	return s.GeneratePair(stored.UserID)
}

// JWT токен
func (s *TokenService) generateAccess(userID uint) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// generateRefresh создаёт случайный токен, сохраняя хеш в БД
func (s *TokenService) generateRefresh(userID uint) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	token := hex.EncodeToString(b)

	rt := &models.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}
	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return "", fmt.Errorf("save refresh token: %w", err)
	}

	return token, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
