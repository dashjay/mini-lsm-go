package block

import "encoding/binary"

type BlockBuilder struct {
	offsets   []uint16
	data      []byte
	blockSize uint64
}

func NewBlockBuilder(size uint64) *BlockBuilder {
	return &BlockBuilder{
		offsets:   make([]uint16, 0),
		data:      make([]byte, 0),
		blockSize: size,
	}
}

func (b *BlockBuilder) estimatedSzie() uint64 {
	return uint64(len(b.offsets))*SizeOfUint16 + uint64(len(b.data)) + SizeOfUint16
}

func (b *BlockBuilder) isEmpty() bool {
	return len(b.offsets) == 0
}

func (b *BlockBuilder) Add(Key, Value string) bool {
	if Key == "" {
		panic("key must not be empty")
	}
	if b.estimatedSzie()+uint64(len(Key))+uint64(len(Value))+
		SizeOfUint16*2+SizeOfUint16 > b.blockSize &&
		!b.isEmpty() {
		return false
	}
	b.offsets = append(b.offsets, uint16(len(b.data)))
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(Key)))
	b.data = append(b.data, Key...)
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(Value)))
	b.data = append(b.data, Value...)
	return true
}

func (b *BlockBuilder) AddByte(Key, Value []byte) bool {
	if len(Key) == 0 {
		panic("key must not be empty")
	}
	if b.estimatedSzie()+uint64(len(Key))+uint64(len(Value))+
		SizeOfUint16*2+SizeOfUint16 > b.blockSize &&
		!b.isEmpty() {
		return false
	}
	b.offsets = append(b.offsets, uint16(len(b.data)))
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(Key)))
	b.data = append(b.data, Key...)
	b.data = binary.BigEndian.AppendUint16(b.data, uint16(len(Value)))
	b.data = append(b.data, Value...)
	return true
}

func (b *BlockBuilder) Build() *Block {
	if b.isEmpty() {
		panic("block should not be empty")
	}
	return &Block{
		data:    b.data,
		offsets: b.offsets,
	}
}
