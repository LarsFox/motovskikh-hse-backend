package main

import (
	"log"

	"github.com/vrischmann/envconfig"

	"github.com/LarsFox/motovskikh-hse-backend/api"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
	"github.com/LarsFox/motovskikh-hse-backend/mysql"
)

// Version — версия приложения.
// Пустое значение переменной подменяется с помощью флага -ldflags во время сборки.
var Version string

// без mysql.Config с ними че-то не то
type config struct {
	Web *webConfig
}

type webConfig struct {
	Addr string `envconfig:"default=:1543"`
}

func main() {
	log.Printf("Version %s", Version)

	// только Web конфиг через envconfig
	cfg := &config{
		Web: &webConfig{},
	}

	// только Web часть
	if err := envconfig.InitWithPrefix(cfg.Web, "motovskikh"); err != nil {
		log.Printf("envconfig warning: %v", err)
	}

	// порт по умолчанию если не прочитался
	if cfg.Web.Addr == "" {
		cfg.Web.Addr = ":1543"
	}

	dbConfig := &mysql.Config{
		User:    "root",   // из .env
		Pass:    "root",   // из .env
		Name:    "hse_db", // из .env
		Host:    "127.0.0.1:3306",
		MaxConn: 8,
	}

	// На всех проверках я подразумеваю, что сайт должен работать,
	// даже если клиент не инициализировался.
	//
	// Если нет БД, лучше отдать индекс.хтмл, чем 500.
	dbClient, err := mysql.NewClient(dbConfig)
	check(err)

	publicAPIManager := api.NewManager(
		manager.New(dbClient),
	)

	check(publicAPIManager.Listen(cfg.Web.Addr))
}

func check(err error) {
	if err == nil {
		return
	}

	log.Printf("cmd main.go init fail: %v", err)
}
