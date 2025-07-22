package lfu

import "container/list"

func (l *lfuStrategy[K]) bumpFreq(n *lfuNode[K]) {
	oldFreq := n.freq
	lst := l.freqLists[oldFreq]
	lst.Remove(n.el)
	if lst.Len() == 0 {
		delete(l.freqLists, oldFreq)
		if oldFreq == l.minFreq {
			l.recalcMinFreq()
		}
	}
	newFreq := oldFreq + 1
	n.freq = newFreq
	newList, ok := l.freqLists[newFreq]
	if !ok {
		newList = list.New()
		l.freqLists[newFreq] = newList
	}
	n.el = newList.PushFront(n.key)
}
