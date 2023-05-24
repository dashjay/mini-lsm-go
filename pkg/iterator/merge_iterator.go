package iterator

import (
	"bytes"
	"log"
)

type MergeIterator struct {
	iters   []Iter
	current int
}

func NewmergeIterator(in ...Iter) *MergeIterator {
	if len(in) == 0 {
		return &MergeIterator{iters: in, current: -1}
	}

	iters := make([]Iter, 0)
	for i := range in {
		if in[i].IsValid() == false {
			continue
		}
		iters = append(iters, in[i])
	}

	for i := range iters {
		log.Printf("%d, NewmergeIterator, next: %s", i, iters[i].Key())
	}
	return &MergeIterator{iters: iters, current: findMinimalIter(iters)}
}

func findMinimalIter(iters []Iter) int {
	// we need to find the least key(the first for duplicate keys)
	min := 0
	for i := 1; i < len(iters); i++ {
		// min.key > i.key
		if bytes.Compare(iters[min].Key(), iters[i].Key()) == 1 {
			min = i
		}
	}
	return min
}

func (m *MergeIterator) Key() []byte {
	return m.iters[m.current].Key()
}

func (m *MergeIterator) Value() []byte {
	return m.iters[m.current].Value()
}

func (m *MergeIterator) IsValid() bool {
	return m.current >= 0 && m.current < len(m.iters) && m.iters[m.current].IsValid()
}

func (m *MergeIterator) Next() {
	currentKey := m.iters[m.current].Key()

	// 1. check iter for current ptr
	m.iters[m.current].Next()
	if !m.iters[m.current].IsValid() {
		m.iters = append(m.iters[:m.current], m.iters[m.current+1:]...)
	}

	// 2. remove all dup keys
	for i := 0; i < len(m.iters); i++ {
		if bytes.Equal(m.iters[i].Key(), currentKey) {
			m.iters[i].Next()
		}
	}

	// 3. remove all invalid
	i := len(m.iters) - 1
	for i >= 0 {
		if !m.iters[i].IsValid() {
			m.iters = append(m.iters[:i], m.iters[i+1:]...)
		}
		i--
	}

	m.current = findMinimalIter(m.iters)
}
