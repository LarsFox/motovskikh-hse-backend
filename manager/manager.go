package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/internal/repository"
	"github.com/LarsFox/motovskikh-hse-backend/internal/services"
)

type db interface {
	Stub() bool
}

type Manager struct {
	db           db
	authService  *services.AuthService
	tokenService *services.TokenService
}

func New(
	db db,
	userRepo repository.UserRepository,
	codeRepo repository.VerificationCodeRepository,
	refreshRepo repository.RefreshTokenRepository,
	emailSender services.EmailSender,
	jwtSecret string,
) *Manager {
	tokenService := services.NewTokenService(jwtSecret, refreshRepo)
	authService := services.NewAuthService(userRepo, codeRepo, emailSender)

	return &Manager{
		db:           db,
		authService:  authService,
		tokenService: tokenService,
	}
}

func (m *Manager) Stub() bool {
	return m.db.Stub()
}

func (m *Manager) Auth() *services.AuthService {
	return m.authService
}

func (m *Manager) Token() *services.TokenService {
	return m.tokenService
}
