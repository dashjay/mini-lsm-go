package block

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/dashjay/mini-lsm-go/pkg/utils"
)

var ErrInvalidBlockMeta = errors.New("invalid block meta")

type Meta struct {
	Offset   uint32
	FirstKey []byte
}

// EncodedBlockMeta help append all metaData to bytes buffer
func EncodedBlockMeta(metaList []*Meta) []byte {
	estimateMetadataSize := uint16(0)
	for _, meta := range metaList {
		estimateMetadataSize += SizeOfUint32
		estimateMetadataSize += SizeOfUint16
		estimateMetadataSize += uint16(len(meta.FirstKey))
	}

	var buffer bytes.Buffer
	var buf [SizeOfUint32]byte
	for _, meta := range metaList {
		binary.BigEndian.PutUint32(buf[:SizeOfUint32], meta.Offset)
		buffer.Write(buf[:SizeOfUint32])
		binary.BigEndian.PutUint16(buf[:SizeOfUint16], uint16(len(meta.FirstKey)))
		buffer.Write(buf[:SizeOfUint16])
		buffer.Write(meta.FirstKey)
	}
	utils.Assertf(estimateMetadataSize == uint16(buffer.Len()),
		"buf size error after encoding, estimateMetadataSize: %d should be equal to buffer.Len(): %d", estimateMetadataSize, buffer.Len())

	return buffer.Bytes()
}

// DecodeBlockMeta read []*Meta from byte slice
func DecodeBlockMeta(input []byte) []*Meta {
	var metas = make([]*Meta, 0)
	for len(input) > 0 {
		offset := binary.BigEndian.Uint32(input[:SizeOfUint32])
		input = input[SizeOfUint32:]
		firstKeyLen := binary.BigEndian.Uint16(input[:SizeOfUint16])
		input = input[SizeOfUint16:]
		key := input[:firstKeyLen]
		input = input[firstKeyLen:]
		metas = append(metas, &Meta{Offset: offset, FirstKey: key})
	}
	return metas
}

func readUint32(r io.Reader) (uint32, error) {
	var temp [SizeOfUint32]byte
	n, err := r.Read(temp[:])
	if err != nil {
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, ErrInvalidBlockMeta
	}
	if uint16(n) != SizeOfUint32 {
		return 0, ErrInvalidBlockMeta
	}
	return binary.BigEndian.Uint32(temp[:]), nil
}

func readUint16(r io.Reader) (uint16, error) {
	var temp [SizeOfUint16]byte
	n, err := r.Read(temp[:])
	if err != nil {
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, ErrInvalidBlockMeta
	}
	if uint16(n) != SizeOfUint16 {
		return 0, ErrInvalidBlockMeta
	}
	return binary.BigEndian.Uint16(temp[:]), nil
}

// DecodeBlockMetaFromReader reads []*Meta from reader
func DecodeBlockMetaFromReader(r io.Reader) ([]*Meta, error) {
	var metas = make([]*Meta, 0)
	for {
		offset, err := readUint32(r)
		if err != nil {
			if err == io.EOF {
				return metas, nil
			}
			return nil, err
		}
		firstKeyLen, err := readUint16(r)
		if err != nil {
			return nil, err
		}
		key := make([]byte, firstKeyLen)
		n, err := r.Read(key)
		if err != nil {
			return nil, err
		}
		if n != int(firstKeyLen) {
			return nil, ErrInvalidBlockMeta
		}
		metas = append(metas, &Meta{Offset: offset, FirstKey: key})
	}
}
