package mp

const privateRoomNameLen = 10

func isPrivateRoom(hash string) bool {
	return len([]rune(hash)) > privateRoomNameLen
}
