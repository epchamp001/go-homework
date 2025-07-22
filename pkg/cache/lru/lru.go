package lru

import (
	"container/list"
	"sync"
)

// lruStrategy реализует классический LRU поверх container/list.
// K хранится как значение в узле list; map хранит ссылку на узел.
type lruStrategy[K comparable] struct {
	cap int
	lst *list.List // MRU — спереди, LRU — сзади
	idx map[K]*list.Element
	mu  sync.Mutex
}

// NewLRUStrategy создаёт стратегию LRU для заданной ёмкости.
func NewLRUStrategy[K comparable](capacity int) *lruStrategy[K] {
	return &lruStrategy[K]{
		cap: capacity,
		lst: list.New(),
		idx: make(map[K]*list.Element),
	}
}
