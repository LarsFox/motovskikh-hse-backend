package mysql

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config — конфигурация клиента.
type Config struct {
	Host    string `envconfig:"optional,default=localhost:3306"`
	Pass    string `envconfig:"MOTOVSKIKH_DB_PASS"`
	MaxConn int    `envconfig:"default=8"`
	Name    string `envconfig:"MOTOVSKIKH_DB_NAME"`
	User    string `envconfig:"MOTOVSKIKH_DB_USER"`
}

func (c *Config) connection() string {
	sqlHost := c.Host
	if !strings.Contains(sqlHost, "tcp") {
		sqlHost = fmt.Sprintf("tcp(%s)", sqlHost)
	}
	return fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Pass, sqlHost, c.Name)
}

type Client struct {
	db *gorm.DB
}

func NewClient(cfg *Config) (*Client, error) {
	db, err := gorm.Open(mysql.Open(cfg.connection()))
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	d, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	// db = db.Debug()
	d.SetMaxOpenConns(cfg.MaxConn)
	return &Client{db: db}, nil
}
