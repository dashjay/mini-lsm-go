package lsm_test

import (
	"testing"

	"github.com/dashjay/mini-lsm-go/pkg/lsm"
	"github.com/dashjay/mini-lsm-go/pkg/test"
	"github.com/stretchr/testify/assert"
)

func genKv(keycount int, dir string) *lsm.Storage {
	lsmKV := lsm.NewStorage(dir)
	const count = 1000
	for i := uint64(0); i < count; i++ {
		lsmKV.Put(test.KeyOf(i), test.ValueOf(i))
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
	scanner := lsmKV.Scan(test.KeyOf(500), test.KeyOf(509))
	for i := uint64(500); i < 510; i++ {
		assert.Equal(t, scanner.IsValid(), true)
		assert.Equal(t, test.KeyOf(i), scanner.Key())
		assert.Equal(t, test.ValueOf(i), scanner.Value())
		scanner.Next()
	}
}
