// Package cache предоставляет потокобезопасный in-memory кэш с обобщёнными
// типами ключей и значений, TTL-просрочкой и настраиваемой стратегией
// выселения (LRU/LFU).
package cache

import (
	"pvz-cli/pkg/cache/lru"
	"sync"
	"time"
)

// Config задаёт параметры кэша.
//   - Capacity   - максимальное количество элементов (0 = без лимита).
//   - TTL        - срок жизни значения с момента последнего Set.
//   - Strategy   - стратегия выселения; если nil и Capacity>0, используется LRU.
//   - OnEvict    - необязательный колбек при удалении пары key/value.
//
// Capacity и Strategy должны быть согласованы: если Capacity==0, Strategy
// игнорируется.
//
// K должен быть comparable, V - произвольный.
type Config[K comparable] struct {
	Capacity int
	TTL      time.Duration
	Strategy EvictStrategy[K]
	OnEvict  func(k K, v any)
}

// Cache - потокобезопасный in-memory кэш.
// NOTE: Дублирование ключей не допускается: Set переопределяет значение.
type Cache[K comparable, V any] struct {
	mu      sync.RWMutex
	items   map[K]entry[V]
	cfg     Config[K]
	closing chan struct{}
}

type entry[V any] struct {
	val V
	exp time.Time
}

// New создаёт Cache по конфигурации.
func New[K comparable, V any](cfg Config[K]) *Cache[K, V] {
	if cfg.Capacity < 0 {
		cfg.Capacity = 0
	}
	if cfg.TTL < 0 {
		cfg.TTL = 0
	}
	if cfg.Strategy == nil && cfg.Capacity > 0 {
		cfg.Strategy = lru.NewLRUStrategy[K](cfg.Capacity)
	}
	c := &Cache[K, V]{
		items:   make(map[K]entry[V]),
		cfg:     cfg,
		closing: make(chan struct{}),
	}
	// Стартуем фонового сборщика просроченного, если задан TTL.
	if cfg.TTL > 0 {
		go c.reaper()
	}
	return c
}
