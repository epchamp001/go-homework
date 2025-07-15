package lfu

func (l *lfuStrategy[K]) RecordDeletion(k K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n, ok := l.nodes[k]
	if !ok {
		return
	}
	lst := l.freqLists[n.freq]
	lst.Remove(n.el)
	delete(l.nodes, k)
	if lst.Len() == 0 {
		delete(l.freqLists, n.freq)
		if n.freq == l.minFreq {
			l.recalcMinFreq()
		}
	}
}
