package manager

type db interface {
	Stub() bool
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{
		db: db,
	}
}

func (m *Manager) Stub() bool {
	return m.db.Stub()
}
