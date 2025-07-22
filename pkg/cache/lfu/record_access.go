package lfu

func (l *lfuStrategy[K]) RecordAccess(k K) {
	l.mu.Lock()
	if n, ok := l.nodes[k]; ok {
		l.bumpFreq(n)
	}
	l.mu.Unlock()
}
