package base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFmtScore(t *testing.T) {
	room := NewRoom("url", "url/#hash", "host", false)
	require.Equal(t, "8", room.FmtScoreFunc(8.00))
	require.Equal(t, "0", room.FmtScoreFunc(.43))
	require.Equal(t, "0", room.FmtScoreFunc(0))
	require.Equal(t, "1", room.FmtScoreFunc(1))
	require.Equal(t, "−12", room.FmtScoreFunc(-12))
}

// nolint:funlen
func TestBaseRoom(t *testing.T) {
	bibaID := "biba"
	bibaNick := "bibabiba"
	pipaID := "pipa"
	pipaNick := "pipapipa"
	kikaID := "kika"
	kikaNick := "kikakika"

	url := "url/#hash"
	host := bibaID
	room := NewRoom("url", url, host, false)

	// Заходит биба, жмет готов.
	newPlayer, isSpec := room.AddPlayer(bibaID, bibaNick, "")
	require.False(t, room.Ready(bibaID))
	require.True(t, newPlayer)
	require.False(t, isSpec)

	// Заходит пипа и кика.
	newPlayer, isSpec = room.AddPlayer(pipaID, pipaNick, "")
	require.True(t, newPlayer)
	require.False(t, isSpec)

	newPlayer, isSpec = room.AddPlayer(kikaID, kikaNick, "")
	require.True(t, newPlayer)
	require.False(t, isSpec)

	room.RenamePlayer(kikaID, bibaNick)
	require.Equal(t, bibaNick, room.Players[kikaID].Nick)

	require.False(t, room.RecolourPlayer(kikaID, "dazo"))

	// кику кикают.
	kicked, start := room.Kick(pipaID, kikaID)
	require.False(t, kicked)
	require.False(t, start)

	kicked, start = room.Kick(bibaID, kikaID)
	require.True(t, kicked)
	require.False(t, start)

	// Биба переподключается дважды, теперь хост пипа.
	nick, start := room.Disconnect(bibaID)
	require.Empty(t, nick)
	require.False(t, start)

	newPlayer, isSpec = room.AddPlayer(bibaID, bibaNick, "")
	require.True(t, newPlayer)
	require.False(t, isSpec)
	require.Equal(t, pipaID, room.host)

	nick, start = room.Disconnect(bibaID)
	require.Empty(t, nick)
	require.False(t, start)

	newPlayer, isSpec = room.AddPlayer(bibaID, bibaNick, "")
	require.True(t, newPlayer)
	require.False(t, isSpec)
	require.False(t, room.Ready(bibaID))

	// Игра начинается, пипа готов.
	require.True(t, room.Ready(pipaID))
	require.EqualValues(t, RoomStateStarted, room.State)

	require.False(t, room.AllowUpdateSettings(bibaID))
	winner, ok := room.GiveUp(bibaID)
	require.True(t, ok)
	require.Equal(t, winner, pipaID)
}
