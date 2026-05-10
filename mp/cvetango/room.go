package cvetango

import (
	"github.com/LarsFox/motovskikh-hse-backend/mp/base"
)

type Room struct {
	base *base.Room
}

func NewRoom(url, host string, private bool) *Room {
	room := &Room{
		base: base.NewRoom("cvetango", url, host, private),
	}

	room.base.StartFunc = room.newRound
	room.base.ScoreFunc = room.scoreFunc

	return room
}

func (r *Room) Base() *base.Room { return r.base }

func (r *Room) Hit(_ string, _ []int) base.HitType {
	return base.HitTypeCorrect
}

// newRound замешивает новые карты, новую подсказку и новый способ отображения.
func (r *Room) newRound() {}

// scoreFunc возвращает текущий результат игрока.
func (r *Room) scoreFunc(playerName string) float64 {
	p := r.base.Players[playerName]
	if p == nil {
		return 0
	}
	return float64(p.Correct)
}
