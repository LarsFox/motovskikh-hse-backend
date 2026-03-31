package mysql

import (
	"fmt"
	"strings"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type Config struct {
	Host    string `envconfig:"optional"`
	Pass    string
	MaxConn int `envconfig:"default=0"`
	Name    string
	User    string
}

type Client struct {
	db *gorm.DB
}

func (c *Config) connection() string {
	sqlHost := c.Host
	if !strings.Contains(sqlHost, "tcp") {
		sqlHost = fmt.Sprintf("tcp(%s)", sqlHost)
	}
	return fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Pass, sqlHost, c.Name)
}

func NewClient(cfg *Config) (*Client, error) {
	db, err := gorm.Open(mysql.Open(cfg.connection()))
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	// Автомиграция для всех таблиц
	err = db.AutoMigrate(
		&entities.TestBucket{},
	)
	if err != nil {
		return nil, fmt.Errorf("dbs migrate err: %w", err)
	}

	d, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	d.SetMaxOpenConns(cfg.MaxConn)
	return &Client{db: db}, nil
}
