package ws

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
	defaultKeysCap = 20
	socketTTL      = time.Minute * 15
)

type Connector struct {
	m        sync.Map
	upgrader *websocket.Upgrader
}

func New(host string) *Connector {
	local := host == "" || strings.Contains(host, "localhost")
	co := &Connector{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return local || "https://"+r.Host == host
			},
		},
	}
	go co.clean()
	return co
}

func connectionKey(room, player string) string {
	return room + "__" + player
}

// TODO: определить, а нужна ли вообще очистка сокетов.
func (s *Connector) clean() {
	t := time.NewTicker(time.Hour)
	for range t.C {
		s.m.Range(func(key, value any) bool {
			conn, ok := value.(*connection)
			if !ok {
				return true
			}

			conn.RLock()
			defer conn.RUnlock()
			if conn.last.Add(socketTTL).After(time.Now()) {
				return true
			}

			s.m.Delete(key)
			return true
		})
	}
}

func (s *Connector) add(room, player string, conn *websocket.Conn) {
	s.m.Store(connectionKey(room, player), &connection{conn: conn, last: time.Now()})
}

func (s *Connector) Delete(room, player string) {
	key := connectionKey(room, player)
	val, ok := s.m.Load(key)
	if !ok {
		return
	}

	c, ok := val.(*connection)
	if !ok {
		return
	}

	c.conn.Close()
	s.m.Delete(connectionKey(room, player))
}

// Upgrade переводит соединение на вебсокет.
// nolint:ireturn
func (s *Connector) Upgrade(
	w http.ResponseWriter,
	r *http.Request,
	responseHeader http.Header,
	room, player string,
) (entities.Connection, error) {
	conn, err := s.upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}

	s.add(room, player, conn)
	return &connection{conn: conn}, nil
}
