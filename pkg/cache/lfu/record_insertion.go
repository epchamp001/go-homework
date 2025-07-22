package lfu

import "container/list"

func (l *lfuStrategy[K]) RecordInsertion(k K) (evictK K, shouldEvict bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Если уже есть — считаем как access
	if n, exists := l.nodes[k]; exists {
		l.bumpFreq(n)
		var zero K
		return zero, false
	}

	// Если нужно выселять
	if l.cap > 0 && len(l.nodes) >= l.cap {
		// Evict ключ с minFreq (LRU среди них - хвост списка)
		llist := l.freqLists[l.minFreq]
		el := llist.Back()
		evictKey := el.Value.(K)
		llist.Remove(el)
		delete(l.nodes, evictKey)
		if llist.Len() == 0 {
			delete(l.freqLists, l.minFreq)
		}
		shouldEvict = true
		evictK = evictKey
	}

	// Вставляем новый ключ с freq=1
	freq := 1
	if lst, ok := l.freqLists[freq]; ok {
		l.nodes[k] = &lfuNode[K]{key: k, freq: freq, el: lst.PushFront(k)}
	} else {
		lst := list.New()
		l.freqLists[freq] = lst
		l.nodes[k] = &lfuNode[K]{key: k, freq: freq, el: lst.PushFront(k)}
	}
	l.minFreq = 1

	return evictK, shouldEvict
}
