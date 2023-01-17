package iterator_test

import (
	"testing"

	"github.com/dashjay/mini-lsm-go/pkg/iterator"
	"github.com/stretchr/testify/assert"
)

type MockIterator struct {
	Data  []struct{ K, V []byte }
	Index uint64
}

func NewMockIterator(data []struct{ K, V []byte }) *MockIterator {
	return &MockIterator{Data: data, Index: 0}
}

var _ iterator.Iter = (*MockIterator)(nil)

func (m *MockIterator) Key() []byte {
	return m.Data[m.Index].K
}
func (m *MockIterator) Value() []byte {
	return m.Data[m.Index].V
}
func (m *MockIterator) IsValid() bool {
	return m.Index < uint64(len(m.Data))
}
func (m *MockIterator) Next() {
	if m.Index < uint64(len(m.Data)) {
		m.Index += 1
	}
}

func CheckIterResult(t *testing.T, iter iterator.Iter, expected []struct{ K, V []byte }) {
	for i := range expected {
		assert.True(t, iter.IsValid())
		assert.Equal(t, expected[i].K, iter.Key())
		assert.Equal(t, expected[i].V, iter.Value())
		iter.Next()
	}
	assert.False(t, iter.IsValid())
}

func TestTwoMerge1(t *testing.T) {
	i1 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.1")},
		{[]byte("b"), []byte("2.1")},
		{[]byte("c"), []byte("3.1")},
	})
	i2 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.2")},
		{[]byte("b"), []byte("2.2")},
		{[]byte("c"), []byte("3.3")},
		{[]byte("d"), []byte("4.2")},
	})
	CheckIterResult(t, iterator.NewTwoMerger(i1, i2), []struct{ K, V []byte }{
		{[]byte("a"), []byte("1.1")},
		{[]byte("b"), []byte("2.1")},
		{[]byte("c"), []byte("3.1")},
		{[]byte("d"), []byte("4.2")},
	})
}

func TestTwoMerge2(t *testing.T) {
	i1 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.1")},
		{[]byte("b"), []byte("2.1")},
		{[]byte("c"), []byte("3.1")},
		{[]byte("e"), []byte("5.1")},
	})
	i2 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.2")},
		{[]byte("b"), []byte("2.2")},
		{[]byte("c"), []byte("3.3")},
		{[]byte("d"), []byte("4.2")},
	})
	CheckIterResult(t, iterator.NewTwoMerger(i2, i1), []struct{ K, V []byte }{
		{[]byte("a"), []byte("1.2")},
		{[]byte("b"), []byte("2.2")},
		{[]byte("c"), []byte("3.3")},
		{[]byte("d"), []byte("4.2")},
		{[]byte("e"), []byte("5.1")},
	})
}

func TestMerge1(t *testing.T) {
	i1 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.1")},
		{[]byte("b"), []byte("2.1")},
		{[]byte("c"), []byte("3.1")},
	})
	i2 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.2")},
		{[]byte("b"), []byte("2.2")},
		{[]byte("c"), []byte("3.2")},
		{[]byte("d"), []byte("4.2")},
	})
	i3 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("b"), []byte("2.3")},
		{[]byte("c"), []byte("3.3")},
		{[]byte("d"), []byte("4.3")},
	})
	CheckIterResult(t, iterator.NewmergeIterator(i1, i2, i3), []struct{ K, V []byte }{
		{[]byte("a"), []byte("1.1")},
		{[]byte("b"), []byte("2.1")},
		{[]byte("c"), []byte("3.1")},
		{[]byte("d"), []byte("4.2")},
	})
}

func TestMerge2(t *testing.T) {
	i1 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.1")},
		{[]byte("b"), []byte("2.1")},
		{[]byte("c"), []byte("3.1")},
	})
	i2 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("a"), []byte("1.2")},
		{[]byte("b"), []byte("2.2")},
		{[]byte("c"), []byte("3.2")},
		{[]byte("d"), []byte("4.2")},
	})
	i3 := NewMockIterator([]struct{ K, V []byte }{
		{[]byte("b"), []byte("2.3")},
		{[]byte("c"), []byte("3.3")},
		{[]byte("d"), []byte("4.3")},
	})
	CheckIterResult(t, iterator.NewmergeIterator(i3, i2, i1), []struct{ K, V []byte }{
		{[]byte("a"), []byte("1.2")},
		{[]byte("b"), []byte("2.3")},
		{[]byte("c"), []byte("3.3")},
		{[]byte("d"), []byte("4.3")},
	})
}
