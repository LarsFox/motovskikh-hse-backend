package mp

import (
	"sync"

	"github.com/LarsFox/motovskikh-hse-backend/mp/base"
	"github.com/LarsFox/motovskikh-hse-backend/mp/cvetango"
)

type syncStorage struct {
	rooms sync.Map
}

func newStorage() *syncStorage {
	s := &syncStorage{}
	return s
}

type roomable interface {
	Base() *base.Room
}

func (s *syncStorage) Clean() {
	s.rooms.Range(func(key, value any) bool {
		room := value.(roomable)
		b := room.Base()

		if b.IsFresh() {
			return true
		}

		s.rooms.Delete(key)
		return true
	})
}

func (s *syncStorage) Base(name string) *base.Room {
	room := assertType[roomable](s.room(name))
	if room == nil {
		return nil
	}

	return room.Base()
}

// nolint:ireturn
func assertType[T any](data any) T {
	val, ok := data.(T)
	if !ok {
		var notOk T
		return notOk
	}

	return val
}

func (s *syncStorage) room(name string) any {
	value, ok := s.rooms.Load(name)
	if !ok {
		return nil
	}

	return value
}

func (s *syncStorage) Cvetango(name string) *cvetango.Room {
	return assertType[*cvetango.Room](s.room(name))
}

func (s *syncStorage) Save(name string, room roomable) {
	s.rooms.Store(name, room)
}

func (s *syncStorage) total() int64 {
	var i int64
	s.rooms.Range(func(_, _ any) bool {
		i++
		return true
	})

	return i
}
