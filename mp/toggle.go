package mp

func (m *Manager) IsOn() bool { return m.turnedOff.Load() == 0 }
func (m *Manager) TurnOff()   { m.turnedOff.Store(1) }
func (m *Manager) TurnOn()    { m.turnedOff.Store(0) }
