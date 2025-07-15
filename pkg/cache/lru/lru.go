package lru

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

// lruStrategy реализует классический LRU поверх container/list.
// K хранится как значение в узле list; map хранит ссылку на узел.
type lruStrategy[K comparable] struct {
	cap int
	lst *list.List // MRU — спереди, LRU — сзади
	idx map[K]*list.Element
	mu  sync.Mutex
}

// NewLRUStrategy создаёт стратегию LRU для заданной ёмкости.
func NewLRUStrategy[K comparable](capacity int) EvictStrategy[K] {
	return &lruStrategy[K]{
		cap: capacity,
		lst: list.New(),
		idx: make(map[K]*list.Element),
	}
}
