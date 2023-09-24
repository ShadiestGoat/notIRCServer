package main

import "sync"

type Ring[T comparable] struct {
	sync.RWMutex
	items []T
	maxLen int
}

func NewRing[T comparable](max int) *Ring[T] {
	return &Ring[T]{
		RWMutex: sync.RWMutex{},
		items: []T{},
		maxLen: max,
	}
}

func (r *Ring[T]) Insert(v T) {
	r.Lock()
	defer r.Unlock()

	if len(r.items) == r.maxLen {
		r.items = append(r.items[1:], v)
	} else {
		r.items = append(r.items, v)
	}
}

func (r *Ring[T]) Items() []T {
	r.RLock()
	
	s := make([]T, len(r.items))
	copy(s, r.items)
	
	r.RUnlock()
		
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i]
    }
	
	return s
}
