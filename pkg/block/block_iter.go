package block

import (
	"bytes"
	"encoding/binary"
)

type BlockIter struct {
	block *Block
	key   []byte
	value []byte
	idx   uint64
}

func NewBLockIter(block *Block) *BlockIter {
	return &BlockIter{
		block: block,
		key:   make([]byte, 0),
		value: make([]byte, 0),
		idx:   0,
	}
}

func (b *BlockIter) isValid() bool {
	return len(b.key) != 0
}

func (b *BlockIter) SeekToFirst() {
	b.SeekTo(0)
}

func (b *BlockIter) Key() []byte {
	if len(b.key) == 0 {
		panic("invalid iterator")
	}
	return b.key
}

func (b *BlockIter) Value() []byte {
	if len(b.key) == 0 {
		panic("invalid iterator")
	}
	return b.value
}

func (b *BlockIter) SeekTo(idx uint64) {
	if idx >= uint64(len(b.block.offsets)) {
		b.key = nil
		b.value = nil
		return
	}
	offset := uint64(b.block.offsets[idx])
	b.seekToOffset(offset)
	b.idx = idx
}

func (b *BlockIter) Next() {
	b.idx++
	b.SeekTo(b.idx)
}

func (b *BlockIter) SeekToKey(key []byte) {
	low := 0
	high := len(b.block.offsets)

	for low < high {
		mid := (low + (high-low)/2)
		b.SeekTo(uint64(mid))
		if !b.isValid() {
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

func (b *BlockIter) seekToOffset(offset uint64) {
	entry := b.block.data[offset:]
	keyLen := binary.BigEndian.Uint16(entry[:2])

	entry = entry[2:]
	b.key = append(b.key[:0], entry[:keyLen]...)
	entry = entry[keyLen:]

	valueLen := binary.BigEndian.Uint16(entry[:2])
	entry = entry[2:]
	b.value = append(b.value[:0], entry[:valueLen]...)
}
