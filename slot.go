package timing_wheel

import (
	"container/list"
	"sync"
)

type timer struct {
	callback func()
}

type slot struct {
	timers 	*list.List
	sync.Mutex
}

func newSlot() *slot {
	return &slot {
		timers: list.New(),
	}
}

func (s *slot) add(t timer) {
	s.Lock()
	defer s.Unlock()

	s.timers.PushBack(t)
}

func (s *slot) trigger() {
	s.Lock()
	l := s.timers
	s.timers = list.New()
	s.Unlock()

	for t := l.Front(); t != nil; t = t.Next() {
		(t.Value).(timer).callback()
	}
}