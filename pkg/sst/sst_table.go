package sst

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/dashjay/mini-lsm-go/pkg/block"
)

var ErrReadBlockError = errors.New("read block error")

type Table struct {
	fd          *os.File
	metas       []*block.Meta
	metaOffsets uint32
	id          uint32

	// blockCache is a map[[2]uint32]*block.Block
	blockCache sync.Map
}

func OpenTableFromFile(id uint32, blockCache sync.Map, fd *os.File) (*Table, error) {
	fi, err := fd.Stat()
	if err != nil {
		return nil, err
	}
	var rawMetaOffset [block.SizeOfUint32]byte
	fd.ReadAt(rawMetaOffset[:], fi.Size()-block.SizeOfUint32)
	blockMetaOffset := binary.BigEndian.Uint32(rawMetaOffset[:])
	fd.Seek(int64(blockMetaOffset), io.SeekStart)
	rawMetas, err := block.DecodeBlockMetaFromReader(io.LimitReader(fd, fi.Size()-4-int64(blockMetaOffset)))
	return &Table{
		fd:          fd,
		metas:       rawMetas,
		metaOffsets: blockMetaOffset,
		id:          id,
		blockCache:  blockCache,
	}, err
}

func (t *Table) ReadBlock(blockIdx uint32) (*block.Block, error) {
	offset := t.metas[blockIdx].Offset
	var offsetEnd uint32
	if blockIdx < uint32(len(t.metas)-1) {
		offsetEnd = t.metas[blockIdx+1].Offset
	} else {
		offsetEnd = t.metaOffsets
	}
	data := make([]byte, offsetEnd-offset)
	n, err := t.fd.ReadAt(data, int64(offset))
	if err != nil {
		return nil, err
	}
	if n != int(offsetEnd-offset) {
		return nil, ErrReadBlockError
	}
	b := &block.Block{}
	b.Decode(data)
	return b, nil
}

func (t *Table) ReadBlockCached(blockIdx uint32) *block.Block {
	if v, ok := t.blockCache.Load([2]uint32{t.id, blockIdx}); ok {
		return v.(*block.Block)
	}
	blk, err := t.ReadBlock(blockIdx)
	if err != nil {
		log.Printf("ReadBlock error: %s", err)
		return nil
	}
	t.blockCache.Store([2]uint32{t.id, blockIdx}, blk)
	return blk
}

func (t *Table) FindBlockIdx(key []byte) uint32 {
	satSub1 := func(a uint32) uint32 {
		if a > 0 {
			return a - 1
		}
		return 0
	}
	for i := uint32(0); i < t.Len(); i++ {
		// firstKey <= key
		if bytes.Compare(t.metas[i].FirstKey, key) <= 0 {
			return satSub1(i)
		}
	}
	return satSub1(t.Len())
}

func (t *Table) Len() uint32 {
	return uint32(len(t.metas))
}

func (t *Table) Meta() []*block.Meta {
	return t.metas
}
