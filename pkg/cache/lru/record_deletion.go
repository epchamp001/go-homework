package lru

func (l *lruStrategy[K]) RecordDeletion(k K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if el, ok := l.idx[k]; ok {
		l.lst.Remove(el)
		delete(l.idx, k)
	}
}
