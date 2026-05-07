package services

import (
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
	"github.com/LarsFox/motovskikh-hse-backend/internal/repository"
	"github.com/LarsFox/motovskikh-hse-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthService — сервис авторизации
type AuthService struct {
	userRepo    repository.UserRepository
	codeRepo    repository.VerificationCodeRepository
	emailSender EmailSender
}

type EmailSender interface {
	SendVerificationCode(email, code string) error
}

// NewAuthService создаёт новый сервис авторизации
func NewAuthService(
	userRepo repository.UserRepository,
	codeRepo repository.VerificationCodeRepository,
	emailSender EmailSender,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		codeRepo:    codeRepo,
		emailSender: emailSender,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(req *models.CreateUserRequest) error {
	// Валидация email
	if err := utils.ValidateEmail(req.Email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	// Проверка, что email не занят
	existing, _ := s.userRepo.GetByEmail(req.Email)
	if existing != nil {
		return fmt.Errorf("email already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Создаём пользователя в БД
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	// Генерация кода подтверждения
	return s.sendVerificationCode(user.ID, user.Email)
}

// VerifyEmail подтверждает email пользователя
func (s *AuthService) VerifyEmail(req *models.VerifyEmailRequest) error {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	vc, err := s.codeRepo.GetValidCode(user.ID, models.EmailVerification, req.Code)
	if err != nil {
		return fmt.Errorf("invalid or expired code")
	}

	if err := s.codeRepo.MarkAsUsed(vc.ID); err != nil {
		return fmt.Errorf("mark code as used: %w", err)
	}

	return s.userRepo.UpdateEmailVerified(user.ID, true)
}

// SignIn вход в ЛК
func (s *AuthService) SignIn(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.EmailVerified {
		return nil, fmt.Errorf("email not verified")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// ResendCode для повторной отправки кода подтверждения
func (s *AuthService) ResendCode(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	return s.sendVerificationCode(user.ID, user.Email)
}

// sendVerificationCode (приватный метод) генерирует и отправляет код подтверждения
func (s *AuthService) sendVerificationCode(userID uint, email string) error {
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	// Сохраняем код в БД, действует 15 минут
	vc := &models.VerificationCode{
		UserID:    userID,
		Code:      code,
		Type:      models.EmailVerification,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.codeRepo.Create(vc); err != nil {
		return fmt.Errorf("save verification code: %w", err)
	}

	// Отправка кода на email
	return s.emailSender.SendVerificationCode(email, code)
}
