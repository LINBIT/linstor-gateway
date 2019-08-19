package crmcontrol

import "sort"

type IntSet struct {
	set map[int]struct{}
}

func NewIntSet() *IntSet {
	return &IntSet{set: make(map[int]struct{})}
}

func (s *IntSet) Keys() []int {
	var keys []int
	for k := range s.set {
		keys = append(keys, k)
	}

	return keys
}

func (s *IntSet) Add(k int) {
	s.set[k] = struct{}{}
}

func (s *IntSet) Len() int { return len(s.set) }

func (s *IntSet) SortedKeys() []int {
	keys := s.Keys()
	sort.Ints(keys)
	return keys
}

func (s *IntSet) ReverseSortedKeys() []int {
	keys := s.Keys()
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return keys
}

func (s *IntSet) GetFree(min, max int) (int, bool) {
	for i := min; i <= max; i++ {
		if _, ok := s.set[i]; !ok {
			return i, true
		}
	}

	return 0, false
}
