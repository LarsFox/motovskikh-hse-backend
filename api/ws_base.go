package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/mp"
)

type wsIn struct {
	Action string          `json:"a"`
	Data   json.RawMessage `json:"d,omitempty"`
}

// wsBaseRoomHandler — кастомный обработчик комнаты.
type wsBaseRoomHandler struct {
	mp *mp.Manager

	connector connector

	// customFunc обрабатывает все остальные сообщения от игры.
	customFunc func(roomName, player string, message *wsIn)
}

func (m *wsBaseRoomHandler) join(
	w http.ResponseWriter, r *http.Request,
	header http.Header,
	room, player string,
) {
	conn, err := m.connector.Upgrade(w, r, header, room, player)
	if err != nil {
		notify(err)
		return
	}

	defer m.mp.Disconnect(room, player)
	defer m.connector.Delete(room, player)

	for {
		message, err := conn.ReadMessage()
		switch {
		case errors.Is(err, nil):
		case errors.Is(err, entities.ErrNotFound):
			return
		default:
			notify(err)
			return
		}

		go func() {
			defer notifyRecover(map[string]any{
				"ws": "base_handler", "room": room,
			})

			incoming := &wsIn{}
			if err := json.Unmarshal(message, incoming); err != nil {
				return
			}

			m.handle(room, player, "", incoming)
		}()
	}
}

// handle обрабатывает сообщение.
// Если обработчика по умолчанию нет, выполнит кастомный.
func (m *wsBaseRoomHandler) handle(room, player, _ string, message *wsIn) {
	switch message.Action {
	case "hello":
		m.mp.Hello(room, player)
	case "bye":
		m.mp.Disconnect(room, player)
	case "greet":
		m.wsRoomGreet(room, player, message.Data)
	case "join":
		m.wsRoomJoin(room, player, "", message.Data)
	case "kick":
		m.wsRoomKick(room, player, message.Data)
	case "colour":
		m.wsRoomColour(room, player, message.Data)
	case "go":
		m.mp.Ready(room, player)
	case "lonely":
		m.mp.Lonely(room, player)
	case "bla":
		m.wsRoomBla(room, player, message.Data)
	default:
		m.customFunc(room, player, message)
	}
}

func (m *wsBaseRoomHandler) wsRoomBla(room, player string, message json.RawMessage) {
	var blabla string
	if err := json.Unmarshal(message, &blabla); err != nil {
		return
	}

	m.mp.Bla(room, player, blabla)
}

func (m *wsBaseRoomHandler) wsRoomColour(room, player string, message json.RawMessage) {
	var colour string
	if err := json.Unmarshal(message, &colour); err != nil {
		return
	}

	m.mp.Colour(room, player, colour)
}

func (m *wsBaseRoomHandler) wsRoomGreet(room, player string, data json.RawMessage) {
	var nick string
	if err := json.Unmarshal(data, &nick); err != nil {
		return
	}

	m.mp.Rename(room, player, cutNick(nick))
}

func (m *wsBaseRoomHandler) wsRoomJoin(room, player, fill string, data json.RawMessage) {
	var nick string
	if err := json.Unmarshal(data, &nick); err != nil {
		return
	}

	nick = cutNick(nick)
	if nick == "" {
		return
	}

	m.mp.Connect(room, player, nick, fill)
}

func (m *wsBaseRoomHandler) wsRoomKick(room, player string, message json.RawMessage) {
	var nick string
	if err := json.Unmarshal(message, &nick); err != nil {
		return
	}

	m.mp.Kick(room, player, nick)
}
