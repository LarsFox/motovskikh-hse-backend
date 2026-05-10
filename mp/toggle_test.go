package mp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToggle(t *testing.T) {
	m := Manager{}

	require.True(t, m.IsOn())

	m.TurnOff()
	require.False(t, m.IsOn())

	m.TurnOff()
	m.TurnOff()
	require.False(t, m.IsOn())

	m.TurnOn()
	require.True(t, m.IsOn())

	m.TurnOn()
	m.TurnOn()
	require.True(t, m.IsOn())
}
