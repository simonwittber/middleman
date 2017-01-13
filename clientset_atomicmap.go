// Generated by: main
// TypeWriter: atomicmap
// Directive: +gen on ClientSet

package middleman

import (
	"sync"
	"sync/atomic"
)

// ClientSetAtomicMap is a copy-on-write thread-safe map of ClientSet
type ClientSetAtomicMap struct {
	mu  sync.Mutex
	val atomic.Value
}

type _ClientSetMap map[string]ClientSet

// NewClientSetAtomicMap returns a new initialized ClientSetAtomicMap
func NewClientSetAtomicMap() *ClientSetAtomicMap {
	am := &ClientSetAtomicMap{}
	am.val.Store(make(_ClientSetMap, 0))
	return am
}

// Get returns a ClientSet for a given key
func (am *ClientSetAtomicMap) Get(key string) (value ClientSet, ok bool) {
	value, ok = am.val.Load().(_ClientSetMap)[key]
	return value, ok
}

// GetAll returns the underlying map of ClientSet
// this map must NOT be modified, to change the map safely use the Set and Delete
// functions and Get the value again
func (am *ClientSetAtomicMap) GetAll() map[string]ClientSet {
	return am.val.Load().(_ClientSetMap)
}

// Len returns the number of elements in the map
func (am *ClientSetAtomicMap) Len() int {
	return len(am.val.Load().(_ClientSetMap))
}

// Set inserts in the map a ClientSet under a given key
func (am *ClientSetAtomicMap) Set(key string, value ClientSet) {
	am.mu.Lock()
	defer am.mu.Unlock()

	m1 := am.val.Load().(_ClientSetMap)
	m2 := make(_ClientSetMap, len(m1)+1)
	for k, v := range m1 {
		m2[k] = v
	}

	m2[key] = value
	am.val.Store(m2)
	return
}

// Delete removes the ClientSet under key from the map
func (am *ClientSetAtomicMap) Delete(key string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	m1 := am.val.Load().(_ClientSetMap)
	_, ok := m1[key]
	if !ok {
		return
	}

	m2 := make(_ClientSetMap, len(m1)-1)
	for k, v := range m1 {
		if k != key {
			m2[k] = v
		}
	}

	am.val.Store(m2)
	return
}
