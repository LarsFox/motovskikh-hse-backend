package api

import (
	"encoding/json"
	"net/http"
)

type wsInCvetangoHit struct {
	Pick []int `json:"i"`
}

type wsCvetangoHandler struct {
	*wsBaseRoomHandler
}

// Наблюдатели на клиенте игнорируют ненужные события.
func (m *wsCvetangoHandler) wsCvetangoHit(room, player string, message json.RawMessage) {
	in := &wsInCvetangoHit{}
	if err := json.Unmarshal(message, in); err != nil {
		return
	}

	m.mp.CvetangoHit(room, player, in.Pick)
}

func (m *Manager) hndlrCvetangoJoin(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("r")
	if hash == "" {
		return
	}

	header, player, err := getPlayerHeader(r)
	if err != nil {
		notify(err)
		return
	}

	room := m.mp.InitCvetangoRoom(hash, player)

	game := &wsCvetangoHandler{}
	game.wsBaseRoomHandler = &wsBaseRoomHandler{
		mp:         m.mp,
		connector:  m.connector,
		customFunc: game.wsCvetango,
	}

	game.join(w, r, header, room, player)
}

func (m *wsCvetangoHandler) wsCvetango(roomName, playerName string, message *wsIn) {
	switch message.Action {
	case "hit", "miss":
		m.wsCvetangoHit(roomName, playerName, message.Data)
	}
}
