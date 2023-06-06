package block

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/utils"
)

const (
	SizeOfUint16 = uint16(unsafe.Sizeof(uint16(0)))
	SizeOfUint32 = uint16(unsafe.Sizeof(uint32(0)))
)

type Block struct {
	data    []byte
	offsets []uint16
}

func (b *Block) estimateBlockByteSize() uint16 {
	return /* 1. offset */ SizeOfUint16 +
		/* 2. offset items */ uint16(len(b.offsets))*SizeOfUint16 +
		/* 3. data site */ SizeOfUint16 +
		/* 4. data bytes */ uint16(len(b.data))
}

// Encode Block to []byte
func (b *Block) Encode() []byte {
	var buffer bytes.Buffer
	buffer.Grow(int(b.estimateBlockByteSize()))

	offsetLen := len(b.offsets)
	utils.Assertf(offsetLen < math.MaxUint16, "length of data %d should less than 1<<16 - 1", offsetLen)

	var buf [SizeOfUint16]byte
	binary.BigEndian.PutUint16(buf[:], uint16(offsetLen))

	buffer.Write(buf[:])

	for _, offset := range b.offsets {
		binary.BigEndian.PutUint16(buf[:], offset)
		buffer.Write(buf[:])
	}

	dataLen := len(b.data)
	utils.Assertf(dataLen < math.MaxUint16, "length of data %d should less than 1<<16 - 1", dataLen)

	binary.BigEndian.PutUint16(buf[:], uint16(dataLen))
	buffer.Write(buf[:])
	buffer.Write(b.data)

	utils.Assertf(uint16(buffer.Len()) == b.estimateBlockByteSize(),
		"block size should be %d but be %d", b.estimateBlockByteSize(), buffer.Len())
	return buffer.Bytes()
}

// Decode decode Block from []byte
// after return, the in []byte can be release or reuse, we should copy we need from in
func (b *Block) Decode(in []byte) {
	inReader := bytes.NewReader(in)
	offsetsLen, err := readUint16(inReader)
	utils.Assertf(err == nil, "read offset length error: %s", err)

	b.offsets = make([]uint16, offsetsLen)
	for i := uint16(0); i < offsetsLen; i++ {
		b.offsets[i], err = readUint16(inReader)
		utils.Assertf(err == nil, "read offset error: %s", err)
	}

	dataLength, err := readUint16(inReader)
	utils.Assertf(err == nil, "read data size error: %s", err)

	b.data, err = io.ReadAll(inReader)
	utils.Assertf(err == nil, "read data error error: %s", err)
	utils.Assertf(dataLength == uint16(len(b.data)), "block size %d mismatch the recorded size %d", len(b.data), dataLength)
}
