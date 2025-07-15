package lfu

import (
	"container/list"
	"sync"
)

type EvictStrategy[K comparable] interface {
	// RecordAccess вызывается при успешном Get или Set-обновлении.
	RecordAccess(K)
	// RecordInsertion вызывается при добавлении нового ключа. Если shouldEvict true, кэш обязан удалить evictK.
	RecordInsertion(K) (evictK K, shouldEvict bool)
	// RecordDeletion сообщает стратегии, что ключ удалён извне.
	RecordDeletion(K)
}

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
func NewLFUStrategy[K comparable](capacity int) EvictStrategy[K] {
	return &lfuStrategy[K]{
		cap:       capacity,
		freqLists: make(map[int]*list.List),
		nodes:     make(map[K]*lfuNode[K]),
		minFreq:   0,
	}
}
