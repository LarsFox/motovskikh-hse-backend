package mp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LarsFox/motovskikh-hse-backend/mp/cvetango"
)

func TestRoomable(t *testing.T) {
	var a roomable

	c := &cvetango.Room{}

	a = c
	require.NotNil(t, a)
}

func TestStorage(t *testing.T) {
	storage := &syncStorage{}

	storage.Save("cvetango", cvetango.NewRoom("cvetango/#keke", "megaman", false))

	require.NotNil(t, storage.Base("cvetango"))
	require.Nil(t, storage.Base("unknown"))

	require.NotNil(t, storage.Cvetango("cvetango"))
	require.Nil(t, storage.Cvetango("room"))
	require.Nil(t, storage.Cvetango("unknown"))

	require.EqualValues(t, 1, storage.total())
}
