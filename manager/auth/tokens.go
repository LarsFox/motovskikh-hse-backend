package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL    = 15 * time.Minute
	refreshTokenBytes = 32
)

type claims struct {
	jwt.RegisteredClaims

	UserID int64 `json:"user_id"`
}

func GeneratePair(userID int64, secretKey []byte) (*entities.TokenPair, error) {
	accessToken, err := generateAccessToken(userID, secretKey)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &entities.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateAccessToken в ответе за JWT токен.
func generateAccessToken(userID int64, secretKey []byte) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// generateRefreshToken создаёт случайный токен.
func generateRefreshToken() (string, error) {
	b := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	token := hex.EncodeToString(b)

	return token, nil
}

// ValidateAccessToken проверяет access токен и возвращает ID пользователя.
func ValidateAccessToken(tokenString string, secretKey []byte) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		// Перекус таксиста чеееек
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, entities.ErrInvalidInput
		}
		return secretKey, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return 0, entities.ErrInvalidInput
	}

	return c.UserID, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
