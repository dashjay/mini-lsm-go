package sst

import (
	"encoding/binary"
	"os"
	"sync"

	"github.com/dashjay/mini-lsm-go/pkg/block"
)

// TableBuilder can build sst
// 3. save meta for block to metas
type TableBuilder struct {
	// builder is current Block Builder
	builder *block.Builder

	// firstKey: save firstKey for every Block
	firstKey []byte

	// data: append encoded Block to
	data []byte

	// metas saves every meta for built Block
	metas []*block.Meta

	// blockSize is size of every Block
	blockSize uint64
}

func keyDeepcopy(key []byte) []byte {
	out := make([]byte, len(key))
	copy(out, key)
	return out
}

// NewTableBuilder receives max blockSize and return a TableBuilder
func NewTableBuilder(blockSize uint64) *TableBuilder {
	return &TableBuilder{
		builder:   block.NewBlockBuilder(blockSize),
		metas:     make([]*block.Meta, 0),
		blockSize: blockSize,
	}
}

// Add receives a pair of key value(string), if builder has been full, we'll close
// current block, create new Block then add key-value to it.
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

// Add receives a pair of key value([]byte), if builder has been full, we'll close
// current block, create new Block then add key-value to it.
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

// Build build sst with all built block
// WARNING: after Build calling
// the data in TableBuilder is dirty(other metadata was appended to it)
func (t *TableBuilder) Build(id uint32, cache *sync.Map, path string) (*Table, error) {
	t.finishBlock()
	buf := t.data
	metaOffset := uint32(len(buf))
	buf = block.AppendEncodedBlockMeta(t.metas, buf)
	buf = binary.BigEndian.AppendUint32(buf, metaOffset)
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
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
	if !builder.IsEmpty() {
		t.metas = append(t.metas, &block.Meta{
			Offset:   uint32(len(t.data)),
			FirstKey: keyDeepcopy(t.firstKey),
		})
		t.data = append(t.data, builder.Build().Encode()...)
	}
	t.builder = block.NewBlockBuilder(t.blockSize)
}
