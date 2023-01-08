package block

import (
	"encoding/binary"
	"errors"
	"io"
)

var ErrInvalidBlockMeta = errors.New("invalid block meta")

type Meta struct {
	Offset   uint32
	FirstKey []byte
}

func AppendEncodedBlockMeta(metaList []*Meta, buf []byte) []byte {
	estimatedSize := 0
	for _, meta := range metaList {
		estimatedSize += SizeOfUint32
		estimatedSize += SizeOfUint16
		estimatedSize += len(meta.FirstKey)
	}
	originLen := len(buf)
	for _, meta := range metaList {
		buf = binary.BigEndian.AppendUint32(buf, meta.Offset)
		buf = binary.BigEndian.AppendUint16(buf, uint16(len(meta.FirstKey)))
		buf = append(buf, meta.FirstKey...)
	}
	if estimatedSize != len(buf)-originLen {
		panic("buf size error after encoding")
	}
	return buf
}

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
	if n != SizeOfUint32 {
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
	if n != SizeOfUint16 {
		return 0, ErrInvalidBlockMeta
	}
	return binary.BigEndian.Uint16(temp[:]), nil
}

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
