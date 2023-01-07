package block_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/block"
)

var (
	Key1   = []byte("key1")
	Value1 = []byte("value1")

	Key2   = []byte("key2")
	Value2 = []byte("value2")

	Key3   = []byte("key3")
	Value3 = []byte("value3")
)

func randomBlock() *block.Block {
	bb := block.NewBlockBuilder(65535)
	bb.AddByte(Key1, Value1)
	bb.AddByte(Key2, Value2)
	bb.AddByte(Key3, Value3)
	return bb.Build()
}

func TestBlockEncode(t *testing.T) {
	b := randomBlock()
	t.Log(b.Encode())
}

func TestBlockDecode(t *testing.T) {
	b := randomBlock()
	be := b.Encode()
	db := &block.Block{}
	db.Decode(be)
	t.Log(db)
}

func TestBlockIter(t *testing.T) {
	b := randomBlock()
	be := b.Encode()
	db := &block.Block{}
	db.Decode(be)

	iter := block.NewBLockIter(db)

	iter.SeekToFirst()
	if !bytes.Equal(iter.Key(), Key1) || !bytes.Equal(iter.Value(), Value1) {
		t.Error("seek to first error")
	}

	iter.Next()
	if !bytes.Equal(iter.Key(), Key2) || !bytes.Equal(iter.Value(), Value2) {
		t.Error("seek to next error")
	}

	iter.SeekToKey(Key3)
	if !bytes.Equal(iter.Key(), Key3) || !bytes.Equal(iter.Value(), Value3) {
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
