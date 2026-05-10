package mp

import (
	"sort"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/mp/base"
)

const (
	readySetGo       = 3
	mpZzzTimeMinutes = 15
)

type wsRoom struct {
	Colour string         `json:"c"` // Цвет игрока.
	ID     string         `json:"i"` // Айди игрока: у наблюдателя нет цвета, на цвет нельзя завязаться.
	GaveUp bool           `json:"g"` // Сдался ли игрок.
	Host   bool           `json:"h"` // Админ ли игрок.
	Score  []*wsOutPlayer `json:"s"`
}

type wsOutPlayer struct {
	Colour string `json:"c"`
	Fill   string `json:"f,omitempty"` // Цвет ника для премиум челов, #ffdc1b.
	ID     string `json:"i"`
	Host   bool   `json:"h"`
	Nick   string `json:"n"`
	Online bool   `json:"o"`
	Ready  bool   `json:"r"`
	Score  string `json:"s"`

	// Поля для сортировки.
	gaveUp   bool
	endSound string
	joinedAt time.Time
	score    float64
}

func (w wsOutPlayer) isSpectator() bool {
	return base.IsSpectator(w.Colour)
}

type wsOutChat struct {
	Blabla string `json:"b"`
	Colour string `json:"c"`
	Nick   string `json:"n"`
}

type wsOutOnline struct {
	Nick   string `json:"n"`
	Online bool   `json:"o"`
}

func (m *Manager) Bla(roomName, playerName, blabla string) {
	players := m.players(roomName)
	if players == nil {
		return
	}

	for playerID := range players {
		m.sender.Send(roomName, playerID, "bla", &wsOutChat{
			Blabla: blabla,
			Colour: players[playerName].Colour,
			Nick:   players[playerName].Nick,
		})
	}
}

func (m *Manager) Colour(roomName, player, colour string) {
	room := m.storage.Base(roomName)
	if room == nil {
		return
	}

	m.start(room, room.RecolourPlayer(player, colour))
}

func (m *Manager) Connect(roomName, playerName, nick, fill string) {
	room := m.storage.Base(roomName)
	if room == nil {
		return
	}

	// Звук нового игрока важен, только если игра не началась.
	// В запущенную игру заходят только наблюдатели, поэтому и звуки там не нужны.
	newPlayer, isSpectator := room.AddPlayer(playerName, nick, fill)
	if !newPlayer {
		m.sendOnline(roomName, nick, true)
	}

	// Очередной перебор, но можем себе позволить, так как событие важное.
	if newPlayer && !isSpectator {
		for playerID := range m.players(roomName) {
			if playerID != playerName {
				m.sender.Send(roomName, playerID, "sound", "join")
			}
		}
	}

	m.sendPlayers(roomName)
	m.sendSettings(roomName, playerName, room.SettingsFunc())

	if m.Started(roomName) {
		m.sendGo(room, playerName)
	}

	m.sender.Send(roomName, playerName, "sup", room.SupFunc(playerName))
}

func (m *Manager) Disconnect(roomName, player string) {
	room := m.storage.Base(roomName)
	if room == nil {
		return
	}

	nick, started := room.Disconnect(player)
	if nick != "" {
		m.sendOnline(roomName, nick, false)
	}

	m.start(room, started)
}

// Hello не дает соединению прокиснуть пинг-понгом.
func (m *Manager) Hello(room, player string) {
	m.sender.Send(room, player, "moto", nil)
}

func (m *Manager) Kick(roomName, host, player string) {
	room := m.storage.Base(roomName)
	if room == nil {
		return
	}

	kicked, start := room.Kick(host, player)
	if !kicked {
		return
	}

	m.sender.Send(roomName, player, "kick", nil)
	m.start(room, start)
}

func (m *Manager) Lonely(name, player string) {
	if m.sleep(name, player) {
		return
	}

	room := m.storage.Base(name)
	if room == nil {
		return
	}

	m.start(room, room.Lonely(player))
}

func (m *Manager) Ready(roomName, player string) {
	if m.sleep(roomName, player) {
		return
	}

	room := m.storage.Base(roomName)
	if room == nil {
		return
	}

	m.start(room, room.Ready(player))
}

func (m *Manager) Rename(roomName, player, nick string) {
	room := m.storage.Base(roomName)
	if room == nil {
		return
	}

	room.RenamePlayer(player, nick)
	m.sendPlayers(roomName)
}

func (m *Manager) Started(roomName string) bool {
	room := m.storage.Base(roomName)
	if room == nil {
		return false
	}

	room.RLock()
	defer room.RUnlock()
	return room.State == base.RoomStateStarted
}

// Отправляет стандартное сообщение для нежелательного поведения.
func (m *Manager) deny(room, player string) {
	m.sender.Send(room, player, "smart", nil)
}

func (m *Manager) gameTime(roomName string) string {
	room := m.storage.Base(roomName)
	if room == nil {
		return ""
	}

	return room.GameTime()
}

func (m *Manager) players(roomName string) map[string]wsOutPlayer {
	room := m.storage.Base(roomName)
	if room == nil {
		return nil
	}

	list := room.PlayersList()
	players := make(map[string]wsOutPlayer, len(list))
	for playerID, player := range list {
		players[playerID] = wsOutPlayer{
			Colour: player.Colour,
			Fill:   player.Fill,
			Host:   player.Host,
			ID:     playerID,
			Nick:   player.Nick,
			Online: player.Online,
			Ready:  player.Ready,
			Score:  player.Score,

			joinedAt: player.JoinedAt,
			gaveUp:   player.GaveUp,
			score:    player.ScoreVal,
		}
	}

	return players
}

func (m *Manager) sendGG(room string, players map[string]wsOutPlayer) {
	for playerID := range players {
		m.sender.Send(room, playerID, "score", m.gameTime(room))
	}

	// Лишний перебор игроков, но игра уже закончилась, можем позволить.
	lines := m.sendPlayers(room)

	// Финальный аккорд.
	for _, line := range lines {
		if line.endSound == "" {
			continue
		}

		m.sender.Send(room, line.ID, "sound", line.endSound)
	}
}

func (m *Manager) sendGo(room *base.Room, playerName string) {
	m.sender.Send(room.URL, playerName, "go", room.CurrentFunc(playerName))
}

func (m *Manager) sendOnline(room, nick string, online bool) {
	for playerID, p := range m.players(room) {
		if p.Nick == nick {
			continue
		}

		m.sender.Send(room, playerID, "online", &wsOutOnline{
			Nick:   nick,
			Online: online,
		})
	}
}

func (m *Manager) sendPlayers(room string) []*wsOutPlayer {
	players := m.players(room)
	lines := make([]*wsOutPlayer, 0, len(players))
	for _, p := range players {
		lines = append(lines, &p)
	}

	sort.SliceStable(lines, func(i, j int) bool {
		spectatorJ := lines[j].isSpectator()
		if lines[i].isSpectator() != spectatorJ {
			return spectatorJ
		}

		if lines[i].score == lines[j].score {
			return lines[i].joinedAt.Before(lines[j].joinedAt)
		}

		return lines[i].score > lines[j].score
	})

	endSound := "win"
	for _, line := range lines {
		if line.isSpectator() {
			continue
		}

		line.endSound = endSound
		endSound = "lose"
	}

	for playerID, p := range players {
		m.sender.Send(room, playerID, "room", &wsRoom{
			Colour: p.Colour,
			ID:     playerID,
			GaveUp: p.gaveUp,
			Host:   p.Host,
			Score:  lines,
		})
	}

	return lines
}

func (m *Manager) sendSettings(room, player string, settings any) {
	if settings == nil {
		return
	}

	m.sender.Send(room, player, "settings", settings)
}

func (m *Manager) sleep(roomName, player string) bool {
	if m.IsOn() {
		return false
	}

	m.sender.Send(roomName, player, "zzz", mpZzzTimeMinutes)
	return true
}

// start рассылает уведомление об изменении участников и запускает игру.
func (m *Manager) start(room *base.Room, start bool) {
	m.sendPlayers(room.URL)
	if !start {
		return
	}

	room.NewRoundFunc()

	// Надежнее еще раз запросить: вдруг подключился во время обратного отсчета.
	for playerID := range m.players(room.URL) {
		m.sendGo(room, playerID)
	}
}
