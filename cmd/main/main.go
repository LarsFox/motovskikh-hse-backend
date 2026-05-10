package main

import (
	"log"

	"github.com/vrischmann/envconfig"

	"github.com/LarsFox/motovskikh-hse-backend/api"
	"github.com/LarsFox/motovskikh-hse-backend/internal/repository"
	"github.com/LarsFox/motovskikh-hse-backend/internal/services"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
	"github.com/LarsFox/motovskikh-hse-backend/mysql"
)

// Version — версия приложения.
// Пустое значение переменной подменяется с помощью флага -ldflags во время сборки.
var Version string

type config struct {
	DB  *mysql.Config
	Web *webConfig
	JWT *jwtConfig
}

type webConfig struct {
	Addr string `envconfig:"default=:8090"`
}

type jwtConfig struct {
	// Секретный ключ для подписи JWT токенов, храним в святом в .env
	Secret string
}

func main() {
	log.Printf("Version %s", Version)
	cfg := &config{}
	if err := envconfig.InitWithPrefix(cfg, "motovskikh"); err != nil {
		log.Fatal(err)
	}

	// На всех проверках я подразумеваю, что сайт должен работать,
	// даже если клиент не инициализировался.
	//
	// Если нет БД, лучше отдать индекс.хтмл, чем 500.
	dbClient, err := mysql.NewClient(cfg.DB)
	check(err)

	userRepo := repository.NewUserRepository(dbClient.DB())
	codeRepo := repository.NewVerificationCodeRepository(dbClient.DB())
	refreshRepo := repository.NewRefreshTokenRepository(dbClient.DB())

	publicAPIManager := api.NewManager(
		manager.New(
			dbClient,
			userRepo,
			codeRepo,
			refreshRepo,
			&services.FakeEmailSender{},
			cfg.JWT.Secret,
		),
	)

	check(publicAPIManager.Listen(cfg.Web.Addr))
}

func check(err error) {
	if err == nil {
		return
	}

	log.Printf("cmd main.go init fail: %v", err)
}
