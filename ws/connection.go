package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type connection struct {
	sync.RWMutex

	last time.Time
	conn *websocket.Conn
	bb   chan *entities.WSOut // Если не равен нилу, то это ББ, данные отправляются в канал.
}

func (c *connection) ReadMessage() ([]byte, error) {
	mt, message, err := c.conn.ReadMessage()
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
	) {
		return nil, entities.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	if mt == websocket.CloseMessage {
		return nil, entities.ErrNotFound
	}

	return message, nil
}
