package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

//go:generate mockgen -source=manager.go -destination=../generated/mocks/db.go -package=mocks
type db interface {
	GetBucket(testID string) (*entities.TestBucket, error)
	SaveBucket(bucket *entities.TestBucket) error
	GetOrCreateBucket(testID string, questionCount int) (*entities.TestBucket, error)
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{db: db}
}
