package entities

import (
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	playerLetters = "abcdefghijklmnopqrstuvwxyz"
	playerNameLen = 7
	roomNameLen   = 5
)

type Connection interface {
	ReadMessage() ([]byte, error)
}

type CvetangoCard struct {
	Colour string
	Shape  string
	Use    bool
}

type CvetangoSettings struct {
	MissPenalty bool
	ReadySetGo  bool
}

type Player struct {
	Colour   string
	Fill     string
	GaveUp   bool
	Host     bool
	ID       string
	JoinedAt time.Time
	Nick     string
	Online   bool
	Ready    bool
	Score    string
	ScoreVal float64
}

type Room struct {
	Host      string
	MessageID int64
	Name      string
	Players   int64
	Private   bool
	URL       string
}

type Score struct {
	Players []*ScoreLine
	Time    time.Duration
}

type ScoreLine struct {
	Nick     string
	Score    string
	ScoreVal float64
}

type Settings struct {
	Difficulty  string
	MissPenalty bool
	Mode        string
	ReadySetGo  bool
	Speedrun    bool // играем на скорость; на время.
}

type StatsMP struct {
	Name    string
	Private bool
}

func (settings *Settings) Valid() bool {
	switch settings.Difficulty {
	case "normal", "hard", "extreme":
	default:
		return false
	}

	switch settings.Mode {
	case "regions", "capitals", "flags", "arms", "license-plates":
	default:
		return false
	}

	return true
}

func FmtTotalTime(d time.Duration) string {
	t := time.Time{}.Add(d)
	if d.Hours() > 1 {
		return t.Format("15:04:05,00")
	}

	return t.Format("04:05,00")
}

func abc(i int64) string {
	b := make([]byte, i)
	for i := range b {
		b[i] = playerLetters[rand.Int64()%int64(len(playerLetters))]
	}
	return string(b)
}

func NewPlayerName() string {
	return abc(playerNameLen)
}

func NewRoomName() string {
	return abc(roomNameLen)
}

func RoomURL(name, hash string) string {
	return fmt.Sprintf("%s/#%s", name, hash)
}

type WSOut struct {
	Action string `json:"a"`
	Data   any    `json:"d,omitempty"`
}
