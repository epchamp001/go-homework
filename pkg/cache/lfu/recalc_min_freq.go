package lfu

func (l *lfuStrategy[K]) recalcMinFreq() {
	min := 0
	for f := range l.freqLists {
		if min == 0 || f < min {
			min = f
		}
	}
	l.minFreq = min
}
