package lsm_test

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/lsm"
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
	return s2b(fmt.Sprintf("key_%0*d", 8, idx))
}

func ValueOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("value_%0*d", 8, idx))
}

func genKv(keycount int, dir string) *lsm.Storage {
	lsmKV := lsm.NewStorage(dir)
	const count = 1000
	for i := uint64(0); i < count; i++ {
		lsmKV.Put(KeyOf(i), ValueOf(i))
	}
	return lsmKV
}

func TestLSM(t *testing.T) {
	tempDir := t.TempDir()
	kv := genKv(1000, tempDir)

	testRange(t, kv)

	kv.MakeNewMemtable()

	testRange(t, kv)

	assert.Nil(t, kv.SinkImemtableToSST())

	// testRange(t, kv)

	// testRange(t, kv)

	// assert.Nil(t, kv.Sync())

	// testRange(t, kv)

	// kv.Compact()

	// testRange(t, kv)
}

func testRange(t *testing.T, lsmKV *lsm.Storage) {
	scanner := lsmKV.Scan(KeyOf(500), KeyOf(509))
	for i := uint64(500); i < 510; i++ {
		assert.Equal(t, scanner.IsValid(), true)
		assert.Equal(t, KeyOf(i), scanner.Key())
		assert.Equal(t, ValueOf(i), scanner.Value())
		scanner.Next()
	}
}
