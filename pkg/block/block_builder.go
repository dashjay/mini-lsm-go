package block

import (
	"encoding/binary"

	"github.com/dashjay/mini-lsm-go/pkg/utils"
)

// Builder is used to build a block
// implement an Add func for appending key value on Builder's data
// Builder build Block in this layout following:
// data:
// keyLen | key | valueLen | value
type Builder struct {
	offsets   []uint16
	data      []byte
	blockSize uint16
}

// NewBlockBuilder return a Builder for giving size
func NewBlockBuilder(size uint16) *Builder {
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
func (b *Builder) estimatedSize() uint16 {
	return SizeOfUint16 + uint16(len(b.offsets))*SizeOfUint16 + SizeOfUint16 + uint16(len(b.data))
}

func (b *Builder) IsEmpty() bool {
	return len(b.offsets) == 0
}

// Add receives a pair of key value(string), return whether it was added to builder
func (b *Builder) Add(key, value string) bool {
	utils.Assert(key != "", "expect none empty key")

	if b.estimatedSize()+uint16(len(key))+uint16(len(value))+
		SizeOfUint16*2+SizeOfUint16 > b.blockSize &&
		!b.IsEmpty() {
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
	utils.Assert(len(key) != 0, "expect none empty key")

	// estimate size calculate out Block size
	// check if it is enough to append a pair of key, value, their size and an offset.
	if b.estimatedSize()+uint16(len(key))+uint16(len(value))+
		SizeOfUint16*2+SizeOfUint16 > b.blockSize &&
		!b.IsEmpty() {
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
	utils.Assert(!b.IsEmpty(),
		"expect builder is not empty")

	return &Block{
		data:    b.data,
		offsets: b.offsets,
	}
}
