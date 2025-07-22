package lfu

import "container/list"

func (s *lfuStrategy[K]) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.freqLists = make(map[int]*list.List)
	s.nodes = make(map[K]*lfuNode[K])
	s.minFreq = 0
}
