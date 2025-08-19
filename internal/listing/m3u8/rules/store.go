package rules

import (
	"sync"
)

type Store struct {
	channels []*Channel
	counter  int
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		channels: make([]*Channel, 0),
		counter:  0,
	}
}

func (s *Store) Add(channel *Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.channels = append(s.channels, channel)
	s.counter++
}

func (s *Store) All() []*Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.channels
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counter
}
