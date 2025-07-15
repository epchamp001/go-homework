package lru

func (l *lruStrategy[K]) RecordInsertion(k K) (evictK K, shouldEvict bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// уже существует? - просто перемещаем.
	if el, ok := l.idx[k]; ok {
		l.lst.MoveToFront(el)
		var zero K
		return zero, false
	}

	// вставляем новый узел.
	el := l.lst.PushFront(k)
	l.idx[k] = el

	if l.lst.Len() > l.cap {
		// выселить хвост
		tail := l.lst.Back()
		oldK := tail.Value.(K)
		l.lst.Remove(tail)
		delete(l.idx, oldK)
		return oldK, true
	}
	var zero K
	return zero, false
}
