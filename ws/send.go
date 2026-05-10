package ws

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

func (s *Connector) Send(room, player, action string, data any) {
	key := connectionKey(room, player)
	value, ok := s.m.Load(key)
	if !ok {
		return
	}

	out := &entities.WSOut{
		Action: action,
		Data:   data,
	}

	conn := value.(*connection)
	conn.Lock()
	defer conn.Unlock()

	conn.last = time.Now()

	switch err := conn.conn.WriteJSON(out); {
	case errors.Is(err, nil):
	case errors.Is(err, websocket.ErrCloseSent):
	default:
		entities.Notify(err)
	}
}
