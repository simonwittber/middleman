// Generated by: main
// TypeWriter: atomicmap
// Directive: +gen on *Client

package middleman

import (
	"sync"
	"sync/atomic"
)

// ClientAtomicMap is a copy-on-write thread-safe map of pointers to Client
type ClientAtomicMap struct {
	mu  sync.Mutex
	val atomic.Value
}

type _ClientMap map[string]*Client

// NewClientAtomicMap returns a new initialized ClientAtomicMap
func NewClientAtomicMap() *ClientAtomicMap {
	am := &ClientAtomicMap{}
	am.val.Store(make(_ClientMap, 0))
	return am
}

// Get returns a pointer to Client for a given key
func (am *ClientAtomicMap) Get(key string) (value *Client, ok bool) {
	value, ok = am.val.Load().(_ClientMap)[key]
	return value, ok
}

// GetAll returns the underlying map of pointers to Client
// this map must NOT be modified, to change the map safely use the Set and Delete
// functions and Get the value again
func (am *ClientAtomicMap) GetAll() map[string]*Client {
	return am.val.Load().(_ClientMap)
}

// Len returns the number of elements in the map
func (am *ClientAtomicMap) Len() int {
	return len(am.val.Load().(_ClientMap))
}

// Set inserts in the map a pointer to Client under a given key
func (am *ClientAtomicMap) Set(key string, value *Client) {
	am.mu.Lock()
	defer am.mu.Unlock()

	m1 := am.val.Load().(_ClientMap)
	m2 := make(_ClientMap, len(m1)+1)
	for k, v := range m1 {
		m2[k] = v
	}

	m2[key] = value
	am.val.Store(m2)
	return
}

// Delete removes the pointer to Client under key from the map
func (am *ClientAtomicMap) Delete(key string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	m1 := am.val.Load().(_ClientMap)
	_, ok := m1[key]
	if !ok {
		return
	}

	m2 := make(_ClientMap, len(m1)-1)
	for k, v := range m1 {
		if k != key {
			m2[k] = v
		}
	}

	am.val.Store(m2)
	return
}
