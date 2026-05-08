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
	if m.db == nil {
		return false
	}
	return m.db.Stub()
}
