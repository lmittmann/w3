package singleflight

import "sync"

type Map[K comparable, V any] struct {
	mux sync.Mutex
	m   map[K]*call[V]
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		m: make(map[K]*call[V]),
	}
}

type call[V any] struct {
	once sync.Once

	val V
	err error
}

type GetFunc[V any] func() (V, error)

func (c *call[V]) Get(fn GetFunc[V]) (V, error) {
	c.once.Do(func() {
		c.val, c.err = fn()
	})
	return c.val, c.err
}

func (m *Map[K, V]) Do(key K, fn GetFunc[V]) (V, error) {
	m.mux.Lock()
	c, ok := m.m[key]
	if !ok {
		c = &call[V]{}
		m.m[key] = c
	}
	m.mux.Unlock()

	return c.Get(fn)
}

// Clear removes all entries from m, leaving it empty.
func (m *Map[K, V]) Clear() {
	m.mux.Lock()
	defer m.mux.Unlock()

	for key := range m.m {
		delete(m.m, key)
	}
}
