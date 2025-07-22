package lru

func (l *lruStrategy[K]) RecordAccess(k K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if el, ok := l.idx[k]; ok {
		l.lst.MoveToFront(el)
	}
}
