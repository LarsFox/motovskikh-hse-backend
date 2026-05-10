package base

import "time"

const freshTTL = time.Hour * 5

func (r *Room) IsFresh() bool {
	r.RLock()
	defer r.RUnlock()
	return r.last.Add(freshTTL).After(time.Now())
}
