package base

import (
	"time"
)

type Player struct {
	Colour  string // пустой цвет означает наблюдателя.
	Correct int64  // хранит current в игре на время.
	Nick    string
	Score   float64

	fill     string // цвет для премиум игроков.
	joinedAt time.Time
	gaveUp   bool
	online   bool
	ready    bool
}

func (p *Player) IsSpectator() bool {
	return IsSpectator(p.Colour)
}

func (p *Player) Able() bool {
	if p.IsSpectator() || !p.ready {
		return false
	}

	return !p.gaveUp
}
