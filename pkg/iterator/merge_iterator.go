package iterator

import (
	"bytes"
)

// MergeIterator can merge many iters to one
// all different key will remain
// if there are same keys, will take iter which index is small
type MergeIterator struct {
	iters   []Iter
	current int
}

// NewMergeIterator receives one or more iters
// return a MergeIterator
func NewMergeIterator(in ...Iter) *MergeIterator {
	if len(in) == 0 {
		return &MergeIterator{iters: in, current: -1}
	}

	iters := make([]Iter, 0)
	for i := range in {
		if !in[i].IsValid() {
			continue
		}
		iters = append(iters, in[i])
	}

	return &MergeIterator{iters: iters, current: findMinimalIter(iters)}
}

func findMinimalIter(iters []Iter) int {
	// every iter is valid, we want to find the smallest key
	min := 0
	for i := 1; i < len(iters); i++ {
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

// Next should skip all same key in every ite
func (m *MergeIterator) Next() {
	currentKey := make([]byte, len(m.iters[m.current].Key()))
	copy(currentKey, m.iters[m.current].Key())

	// 1. move current iter to next
	m.iters[m.current].Next()
	if !m.iters[m.current].IsValid() {
		m.iters = append(m.iters[:m.current], m.iters[m.current+1:]...)
	}

	// 2. remove all dup keys
	for i := 0; i < len(m.iters); i++ {
		for m.iters[i].IsValid() && bytes.Equal(m.iters[i].Key(), currentKey) {
			m.iters[i].Next()
		}
	}

	// 3. remove all invalid iter
	i := len(m.iters) - 1
	for i >= 0 {
		if !m.iters[i].IsValid() {
			m.iters = append(m.iters[:i], m.iters[i+1:]...)
		}
		i--
	}

	m.current = findMinimalIter(m.iters)
}
