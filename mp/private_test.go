package mp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPrivateRoom(t *testing.T) {
	require.True(t, isPrivateRoom("morethan10!"))
	require.False(t, isPrivateRoom("nope"))
}
