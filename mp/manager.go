package mp

import (
	"sync/atomic"
	"time"
)

type Manager struct {
	pool      *pool
	sender    sender
	storage   *syncStorage
	turnedOff atomic.Int32
}

type sender interface {
	Send(room, player, action string, data any)
}

func New(sender sender) *Manager {
	m := &Manager{
		pool:    newPool(),
		sender:  sender,
		storage: newStorage(),
	}

	go m.clean()

	return m
}

func (m *Manager) clean() {
	go func() {
		t := time.NewTicker(time.Hour)
		for range t.C {
			m.pool.clean()
			m.storage.Clean()
		}
	}()
}

func (m *Manager) Total() int64 {
	return m.storage.total()
}
