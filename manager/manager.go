package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

//go:generate mockgen -source=manager.go -destination=../generated/mocks/db.go -package=mocks
type db interface {
	GetStats(testName string) (*entities.TestStats, error)
	SaveStats(bucket *entities.TestStats) error
	GetOrCreateStats(testName string, questionCount int) (*entities.TestStats, error)
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{db: db}
}
