package base

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

const MaxPlayers = 14

var mpColours = map[string]bool{
	"tomato": true,
	"nice":   true,
	"warm":   true,
	"grass":  true,
	"dazo":   true,
	"violet": true,
	"wine":   true,
	"berry":  true,
	"lime":   true,
	"cold":   true,
	"mint":   true,
	"zefir":  true,
	"tango":  true,
	"clay":   true,
}

var mpColoursOrdered = []string{
	"tomato",
	"nice",
	"warm",
	"grass",
	"dazo",
	"violet",
	"wine",
	"berry",
	"lime",
	"cold",
	"mint",
	"zefir",
	"tango",
	"clay",
}

func fmtRoomScore(score float64) string {
	s := fmt.Sprintf("%.0f", score)
	if score >= 0 {
		return s
	}
	return strings.ReplaceAll(s, "-", "−")
}

func RandomColour() string {
	var c string
	for c = range mpColours {
		return c
	}

	return c
}

func IsSpectator(colour string) bool {
	return colour == ""
}

func Shuffle[T ~[]E, E any](arr T) {
	rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
}
