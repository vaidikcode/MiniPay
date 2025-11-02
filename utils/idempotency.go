package utils

import "sync"

type idempotency struct {
	mu sync.RWMutex
	m  map[string]string
}

var Idem = &idempotency{m: make(map[string]string)}

func (i *idempotency) Get(k string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.m[k]
}

func (i *idempotency) Set(k, v string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.m[k] = v
}
