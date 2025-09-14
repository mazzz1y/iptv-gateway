package rules

type Store struct {
	channels []*Channel
	counter  int
}

func NewStore() *Store {
	return &Store{
		channels: make([]*Channel, 0),
		counter:  0,
	}
}

func (s *Store) Add(channel *Channel) {
	s.channels = append(s.channels, channel)
	s.counter++
}

func (s *Store) Replace(channels []*Channel) {
	s.channels = channels
}

func (s *Store) All() []*Channel {
	return s.channels
}

func (s *Store) Len() int {
	return s.counter
}
