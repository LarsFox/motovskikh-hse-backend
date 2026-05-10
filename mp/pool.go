package mp

import (
	"sync"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/mp/base"
)

const poolInitSize = 50

type pool struct {
	sync.RWMutex

	rooms map[string]map[string]*base.Room
}

func newPool() *pool {
	return &pool{rooms: make(map[string]map[string]*base.Room, poolInitSize)}
}

func (p *pool) add(name, hash string, room *base.Room) {
	p.Lock()
	defer p.Unlock()

	rooms, ok := p.rooms[name]
	if !ok {
		rooms = map[string]*base.Room{}
		p.rooms[name] = rooms
	}

	if _, ok := rooms[hash]; ok {
		return
	}

	rooms[hash] = room
}

func (p *pool) clean() {
	p.Lock()
	defer p.Unlock()

	for _, rooms := range p.rooms {
		for hash, room := range rooms {
			if !room.IsFresh() {
				delete(rooms, hash)
			}
		}
	}
}

func (p *pool) find(name string) string {
	p.Lock()
	defer p.Unlock()

	rooms, ok := p.rooms[name]
	if !ok {
		rooms = map[string]*base.Room{}
		p.rooms[name] = rooms
	}

	for roomID, room := range rooms {
		if room.IsPoolable() {
			return roomID
		}
	}

	return entities.NewRoomName()
}

func (m *Manager) RoomPoolHash(name string) string {
	return m.pool.find(name)
}
