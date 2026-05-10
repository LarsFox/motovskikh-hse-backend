package ws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnector(t *testing.T) {
	c := New("")

	c.add("mama", "player", nil)

	val, ok := c.m.Load(connectionKey("mama", "player"))
	require.True(t, ok)
	require.NotNil(t, val)

	conn, ok := val.(*connection)
	require.True(t, ok)
	require.NotNil(t, conn)
	require.Nil(t, conn.bb)
}
