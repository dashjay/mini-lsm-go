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
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		lsmKV.Put([]byte(key), []byte(value))
	}
	return lsmKV
}

func TestLSM(t *testing.T) {
	tempDir := t.TempDir()
	kv := genKv(1000, tempDir)
	testRange(t, kv)

	assert.Nil(t, kv.Sync())

	// kv.DebugScan([]byte("key-500"), []byte("key-509"))

	testRange(t, kv)

	// for i := 0; i < 1000; i++ {
	// 	kv.Put([]byte(fmt.Sprintf("keyn-%d", i)), []byte(fmt.Sprintf("valuen-%d", i)))
	// }

	// testRange(t, kv)

	// assert.Nil(t, kv.Sync())

	// testRange(t, kv)

	// kv.Compact()

	// testRange(t, kv)
}

func testRange(t *testing.T, lsmKV *lsm.Storage) {
	scanner := lsmKV.DebugScan([]byte("key-500"), []byte("key-509"))
	for i := 500; i < 510; i++ {
		assert.Equal(t, scanner.IsValid(), true)
		assert.Equal(t, string(scanner.Key()), fmt.Sprintf("key-%d", i))
		assert.Equal(t, string(scanner.Value()), fmt.Sprintf("value-%d", i))
		scanner.Next()
	}
}
