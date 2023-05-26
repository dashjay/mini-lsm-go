package block

import (
	"encoding/binary"
)

// Builder is used to build a block
// implement an Add func for appending key value on Builder's data
// Builder build Block in this layout following:
// data:
// keyLen | key | valueLen | value
type Builder struct {
	offsets   []uint16
	data      []byte
	blockSize uint64
}

// NewBlockBuilder return a Builder for giving size
func NewBlockBuilder(size uint64) *Builder {
	return &Builder{
		offsets:   make([]uint16, 0),
		data:      make([]byte, 0),
		blockSize: size,
	}
}

// estimatedSize is for estimateSize for Block
// layout of Block is like this:
// | data(N Byte) | offset0(2 Byte) | offset1 ... | offsetN | offsetNum(dataNum | 2Byte) |
// Builder can estimate size of a Block
func (b *Builder) estimatedSzie() uint64 {
	return uint64(len(b.data)) + uint64(len(b.offsets))*SizeOfUint16 + SizeOfUint16
}

func (b *Builder) isEmpty() bool {
	return len(b.offsets) == 0
}

// Add receives a pair of key value(string), return whether it was added to builder
func (b *Builder) Add(key, value string) bool {
	if key == "" {
		panic("key must not be empty")
	}
	if b.estimatedSzie()+uint64(len(key))+uint64(len(value))+
		SizeOfUint16*2+SizeOfUint16 > b.blockSize &&
		!b.isEmpty() {
		return false
	}
	b.offsets = append(b.offsets, uint16(len(b.data)))
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(key)))
	b.data = append(b.data, key...)
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(value)))
	b.data = append(b.data, value...)
	return true
}

// AddByte receives a pair of key value([]byte), return whether it was added to builder
func (b *Builder) AddByte(key, value []byte) bool {
	if len(key) == 0 {
		panic("key must not be empty")
	}
	// estimate size calculate out Block size
	// check if it is enough to append a pair of key, value, their size and an offset.
	if b.estimatedSzie()+uint64(len(key))+uint64(len(value))+
		SizeOfUint16*2+SizeOfUint16 > b.blockSize &&
		!b.isEmpty() {
		return false
	}
	b.offsets = append(b.offsets, uint16(len(b.data)))
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(key)))
	b.data = append(b.data, key...)
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(value)))
	b.data = append(b.data, value...)
	return true
}

// Build return the Block which Builder built
func (b *Builder) Build() *Block {
	if b.isEmpty() {
		panic("block should not be empty")
	}
	return &Block{
		data:    b.data,
		offsets: b.offsets,
	}
}
