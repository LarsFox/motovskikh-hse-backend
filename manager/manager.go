package manager

import (
	"context"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

//go:generate mockgen -source=manager.go -destination=../generated/mocks/db.go -package=mocks
type db interface {
	GetStats(ctx context.Context, testName string) (*entities.TestStats, error)
	SaveStats(ctx context.Context, stats *entities.TestStats) error
	Stub() bool
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Stub() bool {
	return m.db.Stub()
}