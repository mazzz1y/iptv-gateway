package channel

import "sync"

type Registry struct {
	channels []*Channel
	byID     map[string][]*Channel
	byName   map[string][]*Channel
	counter  int
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		channels: make([]*Channel, 0),
		byID:     make(map[string][]*Channel),
		byName:   make(map[string][]*Channel),
		counter:  0,
	}
}

func (r *Registry) Add(channel *Channel) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.channels = append(r.channels, channel)
	r.counter++

	if id := channel.ID(); id != "" {
		r.byID[id] = append(r.byID[id], channel)
	}

	name := channel.Name()
	r.byName[name] = append(r.byName[name], channel)
}

func (r *Registry) All() []*Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channels
}

func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.counter
}

func (r *Registry) ByID(id string) []*Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byID[id]
}

func (r *Registry) ByName(name string) []*Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byName[name]
}

func (r *Registry) BySubscription(subscription Subscription) []*Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Channel, 0)
	for _, ch := range r.channels {
		if ch.subscription == subscription {
			result = append(result, ch)
		}
	}
	return result
}

func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.channels = r.channels[:0]
	r.byID = make(map[string][]*Channel)
	r.byName = make(map[string][]*Channel)
	r.counter = 0
}
