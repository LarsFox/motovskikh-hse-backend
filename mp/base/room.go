package base

import (
	"sync"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type HitType int64

const (
	HitTypeCorrect HitType = iota + 1
	HitTypeWrong
	HitTypeSpectator
	HitTypeFinal
)

const (
	RoomStateInit = iota + 1
	RoomStateStarted
	RoomStateFinished
)

type Room struct {
	sync.RWMutex

	Players map[string]*Player // map[playerID]*Player
	State   int64
	URL     string // cvetango/#dada, имя-ключ комнаты.

	game       string // russia, cvetango, имя комнаты.
	name       string // название для лобби: бобока.
	host       string // айдишник хоста.
	last       time.Time
	private    bool
	startedAt  time.Time
	totalScore []*totalScore
	totalTime  time.Duration

	FmtScoreFunc func(float64) string        // функция форматирования результата.
	GiveUpFunc   func(winner string)         // функция, когда все сдались.
	ScoreFunc    func(player string) float64 // функция подсчёта результата.
	StartFunc    func()                      // функция сброса и начала новой игры, защищена мьютексом.

	// Функции ниже не используются базовой комнатой.
	// Они участвуют в многопользовательской игре, поэтому прибиваю их к комнате.
	//
	// Методы комнаты не используют мьютекс и поэтому могут использоваться извне.
	// Ответственность за мьютекс лежит на уровне вызова функции.
	CurrentFunc  func(player string) any // возвращает состояние игры для игрока в начале игры и при переподключении к начавшейся.
	NewRoundFunc func()                  // вызывается при начале игры, отправляет обратный отсчет.
	SettingsFunc func() any              // возвращает специфические настройки комнаты.
	SupFunc      func(player string) any // возвращает состояние игры на момент подключения, вызывается единожды.
}

type totalScore struct {
	nick     string
	score    string
	scoreVal float64
}

func NewRoom(game, url, host string, private bool) *Room {
	return &Room{
		host:    host,
		game:    game,
		last:    time.Now(),
		private: private,

		Players:      map[string]*Player{},
		FmtScoreFunc: fmtRoomScore,
		ScoreFunc:    func(string) float64 { return 0 },
		StartFunc:    func() {},
		State:        RoomStateInit,
		URL:          url,

		CurrentFunc:  func(string) any { return nil },
		GiveUpFunc:   func(string) {},
		NewRoundFunc: func() {},
		SettingsFunc: func() any { return nil },
		SupFunc:      func(string) any { return nil },
	}
}

// AddPlayer возвращает два флага: новый ли игрок и наблюдатель ли он.
//
// По умолчанию игрок заходит как игрок.
// Если мест нет, заходит как наблюдатель.
func (r *Room) AddPlayer(name, nick, fill string) (bool, bool) {
	r.Lock()
	defer r.Unlock()

	if r.Players[name] != nil {
		r.Players[name].online = true
		return false, r.Players[name].IsSpectator()
	}

	taken := map[string]bool{}
	for _, player := range r.Players {
		if player.IsSpectator() {
			continue
		}

		taken[player.Colour] = true
	}

	var chosen string
	if r.State != RoomStateStarted {
		for _, colour := range mpColoursOrdered {
			if taken[colour] {
				continue
			}

			chosen = colour
			break
		}
	}

	r.refreshRoomName(name, nick)
	r.Players[name] = &Player{
		Colour: chosen,
		Nick:   nick,

		fill:     fill,
		joinedAt: time.Now(),
		online:   true,
	}

	if r.Players[r.host] == nil || !r.Players[r.host].online {
		r.passHost()
	}

	return true, r.Players[name].IsSpectator()
}

func (r *Room) AllowUpdateSettings(player string) bool {
	r.RLock()
	defer r.RUnlock()
	return r.State != RoomStateStarted && player == r.host
}

// Disconnect отпускает пользователя восвояси.
// Возвращает ник покинувшего для уведомления, если игра началась, и буль, готовы ли все.
func (r *Room) Disconnect(player string) (string, bool) {
	r.Lock()
	defer r.Unlock()

	p := r.Players[player]
	if p == nil {
		return "", false
	}

	p.online = false
	if player == r.host {
		r.passHost()
	}

	if r.State != RoomStateStarted || p.IsSpectator() {
		delete(r.Players, player)
		return "", r.tryStart()
	}

	return p.Nick, false
}

// GameTime возвращает время игры.
// Используется для заполнения финального экрана.
func (r *Room) GameTime() string {
	r.RLock()
	defer r.RUnlock()
	return entities.FmtTotalTime(r.totalTime)
}

// GiveUp отмечает игрока сдавшимся.
// Если все сдались, завершает игру и возвращает идентификатор победителя.
// Если игрок остался один, завершает игру и возвращает его идентификатор.
// Вторым аргументом возвращается флаг, успешно ли сдался игрок.
func (r *Room) GiveUp(playerName string) (string, bool) {
	r.Lock()
	defer r.Unlock()

	p := r.Players[playerName]
	if p == nil {
		return "", false
	}

	if p.IsSpectator() || p.gaveUp {
		return "", false
	}

	p.gaveUp = true

	var online int
	var hasWinner bool
	var winner string
	for key, p := range r.Players {
		if !p.online || p.IsSpectator() {
			continue
		}

		online++
		if p.gaveUp {
			continue
		}

		if hasWinner {
			winner = ""
			continue
		}

		hasWinner = true
		winner = key
	}

	if online == 1 {
		winner = playerName
	}

	if winner == "" {
		return "", true
	}

	r.GiveUpFunc(winner)

	r.GG()
	return winner, true
}

func (r *Room) IsPoolable() bool {
	r.RLock()
	defer r.RUnlock()

	players := r.activePlayers()
	return !r.private && r.State != RoomStateStarted && players > 0
}

// Kick возвращает два буля: удалось ли кикнуть чувака и, если да, началась ли игра.
func (r *Room) Kick(host, player string) (bool, bool) {
	r.Lock()
	defer r.Unlock()

	if r.host != host {
		return false, false
	}

	if r.host == player {
		return false, false
	}

	delete(r.Players, player)

	return true, r.tryStart()
}

// Lonely переводит игрока в готовность и возвращает буль, готовы ли все.
func (r *Room) Lonely(player string) bool {
	r.Lock()
	defer r.Unlock()

	if r.State == RoomStateStarted {
		return false
	}

	p := r.Players[player]
	if p == nil {
		return false
	}

	if p.IsSpectator() {
		return false
	}

	if len(r.Players) > 1 {
		return false
	}

	if player != r.host {
		return false
	}

	r.resetStart()
	return true
}

func (r *Room) PlayersList() map[string]*entities.Player {
	r.RLock()
	defer r.RUnlock()

	players := make(map[string]*entities.Player, len(r.Players))
	for playerID, player := range r.Players {
		players[playerID] = &entities.Player{
			Colour:   player.Colour,
			Fill:     player.fill,
			GaveUp:   player.gaveUp,
			Host:     playerID == r.host,
			ID:       playerID,
			JoinedAt: player.joinedAt,
			Nick:     player.Nick,
			Online:   player.online,
			Ready:    player.ready,
			Score:    r.FmtScoreFunc(player.Score),
			ScoreVal: player.Score,
		}
	}

	return players
}

// Ready переводит игрока в готовность и возвращает буль, готовы ли все.
func (r *Room) Ready(player string) bool {
	r.Lock()
	defer r.Unlock()

	if r.State == RoomStateStarted {
		return false
	}

	p := r.Players[player]
	if p == nil {
		return false
	}

	if p.IsSpectator() {
		return false
	}

	p.ready = !p.ready
	if p.ready {
		return r.tryStart()
	}

	return false
}

// RecolourPlayer меняет цвет игрока.
// Пустой цвет переводит в наблюдатели.
// Если уходит в наблюдатели, а все готовы, начинает игру.
func (r *Room) RecolourPlayer(name, colour string) bool {
	r.Lock()
	defer r.Unlock()

	player := r.Players[name]
	if player == nil {
		return false
	}

	if r.State == RoomStateStarted {
		return false
	}

	// Клиент присылает spectator, но и любой другой «цвет» тоже переведет в наблюдатели.
	if !mpColours[colour] {
		player.Colour = ""
		player.ready = false
		return r.tryStart()
	}

	for id, other := range r.Players {
		if id == name {
			continue
		}

		if other.Colour == colour {
			return false
		}
	}

	player.Colour = colour
	return false
}

func (r *Room) RenamePlayer(name, nick string) {
	r.Lock()
	defer r.Unlock()

	player := r.Players[name]
	if player == nil {
		return
	}

	r.refreshRoomName(name, nick)
	player.Nick = nick
}

func (r *Room) Stats() *entities.StatsMP {
	r.RLock()
	defer r.RUnlock()

	return &entities.StatsMP{
		Name:    r.game,
		Private: r.private,
	}
}

func (r *Room) activePlayers() int64 {
	var players int64
	for _, player := range r.Players {
		if player.online && !player.IsSpectator() {
			players++
		}
	}

	return players
}

func (r *Room) passHost() {
	for newHostID, newHost := range r.Players {
		if newHost.IsSpectator() || !newHost.online {
			continue
		}

		r.host = newHostID
		r.name = newHost.Nick
		return
	}
}

func (r *Room) refreshRoomName(name, nick string) {
	if name != r.host {
		return
	}

	r.name = nick
	r.private = r.private || HasSpam(nick)
}

func (r *Room) resetStart() {
	r.startedAt = time.Now()
	r.State = RoomStateStarted

	for _, player := range r.Players {
		player.Correct = 0
		player.gaveUp = false
		player.Score = 0
	}

	r.StartFunc()
}

// Сбрасывает данные по комнате и игрокам.
func (r *Room) tryStart() bool {
	if len(r.Players) <= 1 {
		return false
	}

	if r.State == RoomStateStarted {
		return false
	}

	var spectators int
	for _, player := range r.Players {
		if player.IsSpectator() {
			spectators++
			continue
		}

		if !player.ready {
			return false
		}
	}

	if spectators == len(r.Players) {
		return false
	}

	r.resetStart()
	return true
}
