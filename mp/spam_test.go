package mp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasSpam(t *testing.T) {
	InitTranslitWords()

	require.True(t, HasSpam("я ебал твой рот"))
	require.True(t, HasSpam("россияхуйня"))
	require.True(t, HasSpam("Кот ебучий. Очень* "))
	require.True(t, HasSpam("ВИТЯ ЗИГОМЕТ"))
	require.True(t, HasSpam("zhe shi megahui"))
	require.True(t, HasSpam("мать твоя шлюха"))
	require.True(t, HasSpam("пизда"))
	require.True(t, HasSpam("я где то далеко в ебенях"))
	require.False(t, HasSpam("ребус"))
	require.False(t, HasSpam("пупсичек"))
	require.True(t, HasSpam("ИЛЬЯ ГУБЕНКО ХУЕСОС"))
	require.False(t, HasSpam("леонид мотовских"))
	require.False(t, HasSpam("надПИСЯми"))
	require.False(t, HasSpam("хлебал твой рот"))
	require.False(t, HasSpam("не надо оскорблять"))
	require.True(t, HasSpam("не надо оскорблядь"))
	require.True(t, HasSpam("америкахуйня"))
	require.True(t, HasSpam("pidors"))
	require.True(t, HasSpam("p1d0r"))
	require.True(t, HasSpam("хуесоc")) // Латиница: *****C
	require.True(t, HasSpam("хуeсоc")) // Латиница: **E**C
	require.True(t, HasSpam("пидopы")) // Латиница: ***OP*
	require.True(t, HasSpam("у3б4н"))
	require.True(t, HasSpam("пид0р"))
}

func BenchmarkHasSpam(_ *testing.B) {
	InitTranslitWords()

	HasSpam("я ебал твой рот")
	HasSpam("надПИСЯми")
	HasSpam("Кот ебучий. Очень* ")
	HasSpam("хлебал твой рот")
	HasSpam("мать твоя шлюха")
	HasSpam("пизда")
	HasSpam("пупсичек")
}
