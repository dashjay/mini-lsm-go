package block_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/block"
	"github.com/stretchr/testify/assert"
)

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func KeyOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("key_%d", idx))
}

func ValueOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("value_%d", idx))
}

func generateBlock(t *testing.T) *block.Block {
	bb := block.NewBlockBuilder(10000)
	for i := uint64(0); i < 100; i++ {
		key := KeyOf(i)
		val := ValueOf(i)
		assert.Equal(t, true, bb.AddByte(key, val))
	}
	return bb.Build()
}

func generateBlockMeta() []*block.Meta {
	var res []*block.Meta
	for i := uint64(0); i < 100; i++ {
		key := KeyOf(i)
		res = append(res, &block.Meta{Offset: uint32(i), FirstKey: key})
	}
	return res
}

func TestGenerateBlock(t *testing.T) {
	generateBlock(t)
}

func TestBlockBuilderMisc(t *testing.T) {
	t.Run("test-block-builderr-single-key", func(t *testing.T) {
		builder := block.NewBlockBuilder(16)
		assert.Equal(t, true, builder.AddByte([]byte("233"), []byte("233333")))
		builder.Build()
	})

	t.Run("test-block-builder-full", func(t *testing.T) {
		builder := block.NewBlockBuilder(16)
		assert.Equal(t, true, builder.AddByte([]byte("11"), []byte("11")))
		assert.Equal(t, false, builder.AddByte([]byte("22"), []byte("22")))
		builder.Build()
	})
	t.Run("test-block-encode", func(t *testing.T) {
		b := generateBlock(t)
		_ = b.Encode()
	})

	t.Run("test-block-decode", func(t *testing.T) {
		b := generateBlock(t)
		be := b.Encode()
		db := &block.Block{}
		db.Decode(be)
		assert.Equal(t, *b, *db)
	})
}
func TestBlockIter(t *testing.T) {
	b := generateBlock(t)
	be := b.Encode()
	db := &block.Block{}
	db.Decode(be)

	iter := block.NewBlockIter(db)

	iter.SeekToFirst()
	key0 := KeyOf(0)
	value0 := ValueOf(0)
	if !bytes.Equal(iter.Key(), key0) || !bytes.Equal(iter.Value(), value0) {
		t.Error("seek to first error")
	}

	iter.Next()
	key1 := KeyOf(1)
	value1 := ValueOf(1)
	if !bytes.Equal(iter.Key(), key1) || !bytes.Equal(iter.Value(), value1) {
		t.Error("seek to next error")
	}

	key50 := KeyOf(50)
	value50 := ValueOf(50)
	iter.SeekToKey(key50)
	if !bytes.Equal(iter.Key(), key50) || !bytes.Equal(iter.Value(), value50) {
		t.Error("seek to key error")
	}
}

func TestBlockMeta(t *testing.T) {
	t.Run("test-block-meta-encode", func(t *testing.T) {
		bms := generateBlockMeta()
		buf := make([]byte, 200)
		rand.Read(buf)
		input := buf
		block.AppendEncodedBlockMeta(bms, input)
		assert.Equal(t, buf, input[:200])
	})

	t.Run("test-block-meta-decode", func(t *testing.T) {
		bms := generateBlockMeta()
		buf := make([]byte, 200)
		rand.Read(buf)
		input := buf
		input = block.AppendEncodedBlockMeta(bms, input)
		assert.Equal(t, buf, input[:200])
		metas := block.DecodeBlockMeta(input[200:])
		assert.Equal(t, len(bms), len(metas))
		assert.Equal(t, bms, metas)
	})
}

func newBlock(blockSize uint64, keyCount int) *block.Block {
	bb := block.NewBlockBuilder(blockSize)
	for i := 0; i < keyCount; i++ {
		bb.Add(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
	}
	return bb.Build()
}

func newBlockIter(blockSize uint64, keyCount int) *block.Iter {
	return block.NewBlockIter(newBlock(blockSize, keyCount))
}

func BenchmarkBlockEncode(b *testing.B) {
	block := newBlock(655350, 10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = block.Encode()
	}
}

func BenchmarkBlockDecode(b *testing.B) {
	blk := newBlock(655350, 10000)
	blockByte := blk.Encode()
	var emptyBlock = &block.Block{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emptyBlock.Decode(blockByte)
	}
}

func BenchmarkBlockBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		newBlockIter(512*1024, 10000)
	}
}

func BenchmarkBlockIter(b *testing.B) {
	count := 1000
	iter := newBlockIter(65535, count)

	b.Run("test seek to idx", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter.SeekTo(uint64(i) % uint64(count))
		}
	})

	b.Run("test seek to key exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			idx := uint64(i) % uint64(count)
			key := s2b(fmt.Sprintf("key-%d", idx))
			iter.SeekToKey(key)
		}
	})

	b.Run("test seek to key not exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			idx := uint64(i)%uint64(count) + uint64(count)
			keyNotExists := s2b(fmt.Sprintf("key-%d", idx))
			iter.SeekToKey(keyNotExists)
		}
	})
}
