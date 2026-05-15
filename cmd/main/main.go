package main

import (
	"log"
	"time"

	"github.com/vrischmann/envconfig"

	"github.com/LarsFox/motovskikh-hse-backend/api"
	"github.com/LarsFox/motovskikh-hse-backend/emailer"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
	"github.com/LarsFox/motovskikh-hse-backend/mysql"
)

// Version — версия приложения.
// Пустое значение переменной подменяется с помощью флага -ldflags во время сборки.
var Version string

const CleanInterval = 24

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

	publicAPIManager := api.NewManager(
		manager.New(
			dbClient,
			&emailer.Client{},
			cfg.JWT.Secret,
		),
	)

	// Очистка истёкших refresh токенов раз в сутки
	go func() {
		ticker := time.NewTicker(CleanInterval * time.Hour)
		for range ticker.C {
			if err := dbClient.DeleteExpired(); err != nil {
				log.Printf("delete expired tokens: %v", err)
			}
		}
	}()

	check(publicAPIManager.Listen(cfg.Web.Addr))
}

func check(err error) {
	if err == nil {
		return
	}

	log.Printf("cmd main.go init fail: %v", err)
}
