package lfu

import (
	"container/list"
	"sync"
)

type lfuNode[K comparable] struct {
	key  K
	freq int
	el   *list.Element // указатель на элемент внутри списка частоты
}

type lfuStrategy[K comparable] struct {
	cap int

	// freq -> список ключей (MRU впереди)
	freqLists map[int]*list.List
	// key -> node
	nodes map[K]*lfuNode[K]

	minFreq int // текущая минимальная частота
	mu      sync.Mutex
}

// NewLFUStrategy создаёт стратегию LFU.
func NewLFUStrategy[K comparable](capacity int) *lfuStrategy[K] {
	return &lfuStrategy[K]{
		cap:       capacity,
		freqLists: make(map[int]*list.List),
		nodes:     make(map[K]*lfuNode[K]),
		minFreq:   0,
	}
}
