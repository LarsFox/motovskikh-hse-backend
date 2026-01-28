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

type config struct {
	Web webConfig
	DB  dbConfig
}

type webConfig struct {
	Addr string `envconfig:"default=:8090"`
}

type dbConfig struct {
	User string `envconfig:"default=root"`
	Pass string `envconfig:"default="`
	Name string `envconfig:"default=hse_db"`
	Host string `envconfig:"default=127.0.0.1:3306"`
}

func main() {
	log.Printf("Version %s", Version)

	var cfg config
	if err := envconfig.InitWithPrefix(&cfg, "MOTOVSKIKH"); err != nil {
		log.Printf("envconfig warning: %v", err)
	}

	dbConfig := &mysql.Config{
		User:    cfg.DB.User,
		Pass:    cfg.DB.Pass,
		Name:    cfg.DB.Name,
		Host:    cfg.DB.Host,
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
