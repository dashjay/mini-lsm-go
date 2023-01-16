package memtable

import (
	"bytes"

	"github.com/dashjay/mini-lsm-go/pkg/sst"
	"github.com/huandu/skiplist"
)

type Table struct {
	m *skiplist.SkipList
}

func NewTable() *Table {
	return &Table{m: skiplist.New(skiplist.Bytes)}
}

func (t *Table) Get(key []byte) []byte {
	ele, ok := t.m.GetValue(key)
	if !ok {
		return nil
	}
	eleValue := ele.([]byte)
	out := make([]byte, len(eleValue))
	copy(out, eleValue)
	return out
}

func inlineDeepcopy(in []byte) (out []byte) {
	out = make([]byte, len(in))
	copy(out, in)
	return out
}

func (t *Table) Put(key, value []byte) {
	t.m.Set(inlineDeepcopy(key), inlineDeepcopy(value))
}

func (t *Table) Scan(lower, upper []byte) *MemTableIterator {
	head := t.m.Find(lower)
	return &MemTableIterator{ele: head, end: upper}
}

func (t *Table) Flush(builder *sst.TableBuilder) {
	head := t.m.Element()
	for {
		builder.AddByte(head.Key().([]byte), head.Value.([]byte))
		next := head.Next()
		if next == nil {
			break
		}
		head = next
	}
}

type MemTableIterator struct {
	ele *skiplist.Element
	end []byte
}

func (m *MemTableIterator) Value() []byte {
	return inlineDeepcopy(m.ele.Value.([]byte))
}

func (m *MemTableIterator) Key() []byte {
	return inlineDeepcopy(m.ele.Key().([]byte))
}

func (m *MemTableIterator) IsValue() bool {
	return m.ele != nil && len(m.ele.Key().([]byte)) != 0
}

func (m *MemTableIterator) Next() {
	m.ele = m.ele.Next()
	if m.ele != nil && bytes.Compare(m.ele.Key().([]byte), m.end) == 1 {
		m.ele = nil
		return
	}
}
