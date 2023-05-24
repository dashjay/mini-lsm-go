package lsm_test

import (
	"fmt"
	"testing"

	"github.com/dashjay/mini-lsm-go/pkg/lsm"
	"github.com/stretchr/testify/assert"
)

func genKv(keycount int, dir string) *lsm.Storage {
	lsmKV := lsm.NewStorage(dir)
	const count = 1000
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key-%0*d", 4, i)
		value := fmt.Sprintf("value-%0*d", 4, i)
		lsmKV.Put([]byte(key), []byte(value))
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

	testRange(t, kv)

	// testRange(t, kv)

	// assert.Nil(t, kv.Sync())

	// testRange(t, kv)

	// kv.Compact()

	// testRange(t, kv)
}

func testRange(t *testing.T, lsmKV *lsm.Storage) {
	scanner := lsmKV.Scan([]byte("key-0500"), []byte("key-0509"))
	for i := 500; i < 510; i++ {
		assert.Equal(t, scanner.IsValid(), true)
		assert.Equal(t, fmt.Sprintf("key-%0*d", 4, i), string(scanner.Key()))
		assert.Equal(t, fmt.Sprintf("value-%0*d", 4, i), string(scanner.Value()))
		scanner.Next()
	}
}
