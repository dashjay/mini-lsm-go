package block

import (
	"encoding/binary"
)

const (
	SizeOfUint16 = 2
	SizeOfUint32 = 4
)

type Block struct {
	data    []byte
	offsets []uint16
}

func (b *Block) Encode() []byte {
	var buf = make([]byte, 0, uint64(len(b.offsets))*SizeOfUint16+uint64(len(b.data))+SizeOfUint16)
	buf = append(buf, b.data...)

	offsetLen := (len(b.offsets))
	for _, offset := range b.offsets {
		buf = binary.BigEndian.AppendUint16(buf, offset)
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(offsetLen))
	return buf
}

func (b *Block) Decode(in []byte) {
	offsets_len := binary.BigEndian.Uint16(in[len(in)-SizeOfUint16:])
	dataEnd := len(in) - SizeOfUint16 - int(offsets_len)*SizeOfUint16
	offsetRaw := in[dataEnd : len(in)-SizeOfUint16]
	b.offsets = make([]uint16, offsets_len)
	for i := uint16(0); i < offsets_len; i++ {
		// i = 0: offsetRaw[0:2]
		// i = 1: offsetRaw[2:4]
		b.offsets[i] = binary.BigEndian.Uint16(offsetRaw[i*2 : i*2+2])
	}
	b.data = in[:dataEnd]
}
