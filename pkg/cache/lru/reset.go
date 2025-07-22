package lru

import "container/list"

func (s *lruStrategy[K]) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lst.Init()
	s.idx = make(map[K]*list.Element)
}
