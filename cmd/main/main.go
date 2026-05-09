package main

import (
	"log"

	"github.com/vrischmann/envconfig"

	"github.com/LarsFox/motovskikh-hse-backend/api"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
	//"github.com/LarsFox/motovskikh-hse-backend/mysql"
)


var Version string

type config struct {
	//DB  *mysql.Config
	Web *webConfig
}

type webConfig struct {
	Addr string `envconfig:"default=:8090"`
}

func main() {
	log.Printf("Version %s", Version)
	cfg := &config{}
	if err := envconfig.InitWithPrefix(cfg, "motovskikh"); err != nil {
		log.Fatal(err)
	}

	publicAPIManager := api.NewManager(
		//manager.New(dbClient),
		manager.New(nil),
	)


	check(publicAPIManager.Listen(cfg.Web.Addr))
}

func check(err error) {
	if err == nil {
		return
	}

	log.Printf("cmd main.go init fail: %v", err)
}
