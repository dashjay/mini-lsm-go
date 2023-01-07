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
}
func TestBlockEncode(t *testing.T) {
	b := generateBlock(t)
	_ = b.Encode()
}

func TestBlockDecode(t *testing.T) {
	b := generateBlock(t)
	be := b.Encode()
	db := &block.Block{}
	db.Decode(be)
}

func TestBlockIter(t *testing.T) {
	b := generateBlock(t)
	be := b.Encode()
	db := &block.Block{}
	db.Decode(be)

	iter := block.NewBLockIter(db)

	iter.SeekToFirst()
	key0 := KeyOf(0)
	value0 := KeyOf(0)
	if !bytes.Equal(iter.Key(), key0) || !bytes.Equal(iter.Value(), value0) {
		t.Error("seek to first error")
	}

	iter.Next()
	key1 := KeyOf(1)
	value1 := KeyOf(1)
	if !bytes.Equal(iter.Key(), key1) || !bytes.Equal(iter.Value(), value1) {
		t.Error("seek to next error")
	}

	key50 := KeyOf(50)
	value50 := KeyOf(50)
	iter.SeekToKey(key50)
	if !bytes.Equal(iter.Key(), key50) || !bytes.Equal(iter.Value(), value50) {
		t.Error("seek to key error")
	}
}

func BenchmarkBlockEncode(b *testing.B) {
	count := 1000

	bb := block.NewBlockBuilder(65535)
	for i := 0; i < count; i++ {
		bb.Add(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = bb.Build().Encode()
	}
	b.StopTimer()
}

func BenchmarkBlockDecode(b *testing.B) {
	count := 1000

	bb := block.NewBlockBuilder(65535)
	for i := 0; i < count; i++ {
		bb.Add(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
	}
	blockByte := bb.Build().Encode()

	var blk = &block.Block{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		blk.Decode(blockByte)
	}
	b.StopTimer()
}

func BenchmarkBlockIter(b *testing.B) {
	count := 1000

	bb := block.NewBlockBuilder(65535)
	for i := 0; i < count; i++ {
		bb.Add(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
	}
	iter := block.NewBLockIter(bb.Build())

	b.Run("test seek to idx", func(b *testing.B) {
		idx := rand.Uint64() % uint64(count)
		iter.SeekTo(idx)
	})

	b.Run("test seek to key exists", func(b *testing.B) {
		idx := rand.Uint64() % uint64(count)
		key := s2b(fmt.Sprintf("key-%d", idx))
		iter.SeekToKey(key)
	})

	b.Run("test seek to key not exists", func(b *testing.B) {
		idx := rand.Uint64()%uint64(count) + uint64(count)
		keyNotExists := s2b(fmt.Sprintf("key-%d", idx))
		iter.SeekToKey(keyNotExists)
	})
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
