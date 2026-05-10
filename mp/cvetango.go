package mp

import (
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/mp/base"
	"github.com/LarsFox/motovskikh-hse-backend/mp/cvetango"
)

type wsCvetangoSettings struct {
	MissPenalty bool `json:"p"`
	ReadySetGo  bool `json:"r"`
}

type wsOutCvetangoHit struct {
	Colour string              `json:"c"`
	Nick   string              `json:"n"`
	Result bool                `json:"r"`
	Round  *wsOutCvetangoRound `json:"d"`
}

type wsOutCvetangoRound struct {
	Mode  string               `json:"m"`
	Round []*wsOutCvetangoCard `json:"d"`
}

type wsOutCvetangoCard struct {
	Colour string `json:"c"`
	Shape  string `json:"s"`
	Use    bool   `json:"u"`
}

func (m *Manager) InitCvetangoRoom(hash, host string) string {
	url := entities.RoomURL("cvetango", hash)
	room := m.storage.Cvetango(url)
	if room != nil {
		return url
	}

	r := cvetango.NewRoom(url, host, isPrivateRoom(hash))

	r.Base().CurrentFunc = func(_ string) any { return m.mapCvetangoRound(url) }
	r.Base().NewRoundFunc = func() { m.sendCvetangoRoomRSG(url) }
	r.Base().SettingsFunc = func() any { return m.cvetangoSettings(url) }

	m.pool.add("cvetango", hash, r.Base())
	m.storage.Save(url, r)

	return url
}

// CvetangoHit пример логики.
func (m *Manager) CvetangoHit(roomName, playerName string, pick []int) {
	// Пробуем найти комнату.
	room := m.storage.Cvetango(roomName)
	if room == nil {
		return
	}

	// Проводим операцию над комнатой.
	room.Base().Lock()
	hit := room.Hit(playerName, pick)
	room.Base().Unlock()

	// Если нажатие неверное, отклоняем.
	switch hit {
	case base.HitTypeWrong, base.HitTypeSpectator:
		m.deny(roomName, playerName) // необязательно.
		return
	}

	// Если верное, уведомляем всех игроков о новом событии.
	players := m.players(roomName)
	for playerID := range players {
		m.sender.Send(roomName, playerID, "hit", &wsOutCvetangoHit{})
	}

	// Если игра завершилась, завершаем игру.
	if hit != base.HitTypeFinal {
		m.sendGG(roomName, players)
	}
}

func (m *Manager) sendCvetangoRoomRSG(room string) {
	if settings := m.cvetangoSettings(room); settings == nil || !settings.ReadySetGo {
		return
	}

	players := m.players(room)
	for i := range readySetGo {
		for playerID := range players {
			m.sender.Send(room, playerID, "rsg", readySetGo-i)
		}

		time.Sleep(time.Second)
	}
}

func (m *Manager) cvetangoSettings(name string) *wsCvetangoSettings {
	room := m.storage.Cvetango(name)
	if room == nil {
		return nil
	}

	room.Base().RLock()
	defer room.Base().RUnlock()
	return &wsCvetangoSettings{}
}

func (m *Manager) mapCvetangoRound(_ string) any {
	return &wsOutCvetangoRound{}
}
