package sst

import (
	"encoding/binary"
	"os"
	"reflect"
	"sync"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/block"
)

type TableBuilder struct {
	builder   *block.Builder
	firstKey  []byte
	data      []byte
	metas     []*block.Meta
	blockSize uint64
}

func keyDeepcopy(key []byte) []byte {
	out := make([]byte, len(key))
	copy(out, key)
	return out
}

func NewTableBuilder(blockSize uint64) *TableBuilder {
	return &TableBuilder{
		builder:   block.NewBlockBuilder(blockSize),
		metas:     make([]*block.Meta, 0),
		blockSize: blockSize,
	}
}

func (t *TableBuilder) Add(key, value string) {
	if t.firstKey == nil {
		t.firstKey = []byte(key)
	}
	if t.builder.Add(key, value) {
		return
	}
	t.finishBlock()
	if !t.builder.Add(key, value) {
		panic("build error")
	}
	t.firstKey = []byte(key)
}

func (t *TableBuilder) AddByte(key, value []byte) {
	if t.firstKey == nil {
		t.firstKey = keyDeepcopy(key)
	}
	if t.builder.AddByte(key, value) {
		return
	}

	t.finishBlock()
	if !t.builder.AddByte(key, value) {
		panic("build error")
	}
	t.firstKey = keyDeepcopy(key)
}

func (t *TableBuilder) Build(id uint32, cache sync.Map, path string) (*Table, error) {
	t.finishBlock()
	buf := t.data
	metaOffset := uint32(len(buf))
	buf = block.AppendEncodedBlockMeta(t.metas, buf)
	buf = binary.BigEndian.AppendUint32(buf, metaOffset)
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	_, err = fd.Write(buf)
	if err != nil {
		return nil, err
	}
	return &Table{
		id:          id,
		fd:          fd,
		metas:       t.metas,
		metaOffsets: metaOffset,
		blockCache:  cache,
	}, nil
}

func (t *TableBuilder) Len() uint32 {
	return uint32(len(t.metas))
}

func (t *TableBuilder) finishBlock() {
	builder := t.builder
	encodedBlock := builder.Build().Encode()
	t.builder = block.NewBlockBuilder(t.blockSize)
	t.metas = append(t.metas, &block.Meta{
		Offset:   uint32(len(t.data)),
		FirstKey: keyDeepcopy(t.firstKey),
	})
	t.data = append(t.data, encodedBlock...)
}

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
