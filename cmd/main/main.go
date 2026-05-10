package main

import (
	"log"

	"github.com/vrischmann/envconfig"

	"github.com/LarsFox/motovskikh-hse-backend/api"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
	"github.com/LarsFox/motovskikh-hse-backend/mp"
	"github.com/LarsFox/motovskikh-hse-backend/mysql"
	"github.com/LarsFox/motovskikh-hse-backend/ws"
)

// Version — версия приложения.
// Пустое значение переменной подменяется с помощью флага -ldflags во время сборки.
var Version string

type config struct {
	DB  *mysql.Config
	Web *webConfig
}

type webConfig struct {
	Addr string `envconfig:"default=:8090"`
	Host string `envconfig:"default=http://localhost:8090"` // motovskikh.ru
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

	connector := ws.New(cfg.Web.Host)
	publicAPIManager := api.NewManager(
		connector,
		manager.New(dbClient),
		mp.New(connector),
	)

	check(publicAPIManager.Listen(cfg.Web.Addr))
}

func check(err error) {
	if err == nil {
		return
	}

	log.Printf("cmd main.go init fail: %v", err)
}
