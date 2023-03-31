package block

import (
	"bytes"
	"encoding/binary"
)

// Iter can hold an Block, for iterating it one-by-one.
// keys in Iter should be sorted
type Iter struct {
	block *Block
	key   []byte
	value []byte
	idx   uint64
}

// NewBlockIter receives a block and return Iter for it.
func NewBlockIter(block *Block) *Iter {
	return &Iter{
		block: block,
		key:   make([]byte, 0),
		value: make([]byte, 0),
		idx:   0,
	}
}

// NewBlockIterAndSeekToFirst receives a block, create a Iter, seek to first key, return it.
func NewBlockIterAndSeekToFirst(block *Block) *Iter {
	i := NewBlockIter(block)
	i.SeekTo(0)
	return i
}

// NewBlockIterAndSeekToKey receives a block, create a Iter, seek to specified key, return it.
func NewBlockIterAndSeekToKey(block *Block, key []byte) *Iter {
	i := NewBlockIter(block)
	i.SeekToKey(key)
	return i
}

// IsValid checks that whether Iter valid
func (b *Iter) IsValid() bool {
	return b.block != nil && len(b.key) != 0
}

// SeekToFirst help Iter to seek to first key
func (b *Iter) SeekToFirst() {
	b.SeekTo(0)
}

// Key get key for current pos
func (b *Iter) Key() []byte {
	if len(b.key) == 0 {
		panic("invalid iterator")
	}
	// WARNING: we assumed that return key will not be modified
	// key := make([]byte, len(b.key))
	// copy(key, b.key)
	// return key
	return b.key
}

// Value get value for current pos
func (b *Iter) Value() []byte {
	if len(b.key) == 0 {
		panic("invalid iterator")
	}
	// WARNING: we assumed that return value will not be modified
	// value := make([]byte, len(b.value))
	// copy(value, b.value)
	// return value
	return b.value
}

// SeekTo receives an index, then try to seek to key-value pair on this index.
func (b *Iter) SeekTo(idx uint64) {
	if b.block == nil {
		return
	}
	if idx >= uint64(len(b.block.offsets)) {
		b.key = nil
		b.value = nil
		return
	}
	offset := uint64(b.block.offsets[idx])
	b.seekToOffset(offset)
	b.idx = idx
}

// Next make iter turn to next key-value pair
func (b *Iter) Next() {
	if b.block == nil {
		return
	}
	b.idx++
	b.SeekTo(b.idx)
}

// SeekToKey make iter to find key in dichotomy.
func (b *Iter) SeekToKey(key []byte) {
	if b.block == nil {
		return
	}
	low := 0
	high := len(b.block.offsets)

	for low < high {
		mid := (low + (high-low)/2)
		b.SeekTo(uint64(mid))
		if !b.IsValid() {
			panic("invalid block")
		}
		switch bytes.Compare(b.key, key) {
		case 0:
			return
		case -1:
			low = mid + 1
		case 1:
			high = mid
		}
	}
	b.SeekTo(uint64(low))
}

func (b *Iter) seekToOffset(offset uint64) {
	entry := b.block.data[offset:]

	keyLen := binary.BigEndian.Uint16(entry[:2])
	entry = entry[2:]
	b.key = append(b.key[:0], entry[:keyLen]...)
	entry = entry[keyLen:]

	valueLen := binary.BigEndian.Uint16(entry[:2])
	entry = entry[2:]
	b.value = append(b.value[:0], entry[:valueLen]...)
}
