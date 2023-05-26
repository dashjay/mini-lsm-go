package lsm

import (
	"testing"

	"github.com/dashjay/mini-lsm-go/pkg/test"
	"github.com/stretchr/testify/assert"
)

func genKv(keyCount uint64, dir string) *StorageInner {
	lsmKV := NewStorageInner(dir)
	for i := uint64(0); i < keyCount; i++ {
		lsmKV.Put(test.KeyOf(i), test.ValueOf(i))
	}
	return lsmKV
}

func TestInternalStorage(t *testing.T) {
	tempDir := t.TempDir()
	kv := genKv(1000, tempDir)

	testRange(t, kv)

	kv.newMemTable()

	testRange(t, kv)

	assert.Nil(t, kv.sinkImMemTableToSST())

	testRange(t, kv)

	kv.Put(test.KeyOf(400), test.ValueOf(0))
	kv.Put(test.KeyOf(401), test.ValueOf(0))

	testRange(t, kv)

	kv.newMemTable()

	testRange(t, kv)

	assert.Nil(t, kv.sinkImMemTableToSST())

	testRange(t, kv)

	kv.compactSSTs()

	testRange(t, kv)
}

func BenchmarkLSM(b *testing.B) {
	lsmKV := NewStorage(b.TempDir())
	for i := 0; i < b.N; i++ {
		key := test.BigKeyOf(uint64(i))
		value := test.BigValueOf(uint64(i))
		lsmKV.Put(key, value)
	}
}

func testRange(t *testing.T, lsmKV *StorageInner) {
	scanner := lsmKV.Scan(test.KeyOf(500), test.KeyOf(509))
	for i := uint64(500); i < 510; i++ {
		assert.Equal(t, scanner.IsValid(), true)
		assert.Equal(t, test.KeyOf(i), scanner.Key())
		assert.Equal(t, test.ValueOf(i), scanner.Value())
		scanner.Next()
	}
}
