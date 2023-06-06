package block_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dashjay/mini-lsm-go/pkg/block"
	"github.com/dashjay/mini-lsm-go/pkg/test"
)

func generateBlock(t *testing.T) *block.Block {
	bb := block.NewBlockBuilder(10000)
	for i := uint64(0); i < 100; i++ {
		key := test.KeyOf(i)
		val := test.ValueOf(i)
		assert.Equal(t, true, bb.AddByte(key, val))
	}
	return bb.Build()
}

func generateBlockMeta() []*block.Meta {
	var res []*block.Meta
	for i := uint64(0); i < 100; i++ {
		key := test.KeyOf(i)
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
	key0 := test.KeyOf(0)
	value0 := test.ValueOf(0)
	assert.Equal(t, key0, iter.Key())
	assert.Equal(t, value0, iter.Value())

	iter.Next()
	key1 := test.KeyOf(1)
	value1 := test.ValueOf(1)
	assert.Equal(t, key1, iter.Key())
	assert.Equal(t, value1, iter.Value())

	key50 := test.KeyOf(50)
	value50 := test.ValueOf(50)
	iter.SeekToKey(key50)
	assert.Equal(t, key50, iter.Key())
	assert.Equal(t, value50, iter.Value())
}

func TestBlockMeta(t *testing.T) {
	t.Run("test-block-meta-encode", func(t *testing.T) {
		bms := generateBlockMeta()
		buf := make([]byte, 200)
		_, _ = rand.Read(buf)
		input := buf
		block.EncodedBlockMeta(bms, input)
		assert.Equal(t, buf, input[:200])
	})

	t.Run("test-block-meta-decode", func(t *testing.T) {
		bms := generateBlockMeta()
		buf := make([]byte, 200)
		_, _ = rand.Read(buf)
		input := buf
		input = block.EncodedBlockMeta(bms, input)
		assert.Equal(t, buf, input[:200])
		metas := block.DecodeBlockMeta(input[200:])
		assert.Equal(t, len(bms), len(metas))
		assert.Equal(t, bms, metas)
	})
}

func newBlock(blockSize uint16, keyCount uint64) *block.Block {
	bb := block.NewBlockBuilder(blockSize)
	for i := uint64(0); i < keyCount; i++ {
		bb.AddByte(test.KeyOf(i), test.ValueOf(i))
	}
	return bb.Build()
}

func newBlockIter(blockSize uint16, keyCount uint64) *block.Iter {
	return block.NewBlockIter(newBlock(blockSize, keyCount))
}

func BenchmarkBlockEncode(b *testing.B) {
	blk := newBlock(4096, 10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = blk.Encode()
	}
}

func BenchmarkBlockDecode(b *testing.B) {
	blk := newBlock(4096, 10000)
	blockByte := blk.Encode()
	var emptyBlock = &block.Block{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emptyBlock.Decode(blockByte)
	}
}

func BenchmarkBlockBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		newBlockIter(4096, 10000)
	}
}

func BenchmarkBlockIter(b *testing.B) {
	count := uint64(1000)
	iter := newBlockIter(65535, count)

	b.Run("test seek to idx", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter.SeekTo(uint64(i) % count)
		}
	})

	b.Run("test seek to key exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			idx := uint64(i) % count
			key := test.KeyOf(idx)
			iter.SeekToKey(key)
		}
	})

	b.Run("test seek to key not exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			idx := uint64(i)%count + count
			keyNotExists := test.KeyOf(idx)
			iter.SeekToKey(keyNotExists)
		}
	})
}
