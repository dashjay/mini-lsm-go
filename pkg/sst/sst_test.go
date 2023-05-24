package sst_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/sst"
	"github.com/stretchr/testify/assert"
)

func TestBuildSSTSingleKey(t *testing.T) {
	tb := sst.NewTableBuilder(16)
	tb.Add("233", "233333")
	tempdir := t.TempDir()
	_, err := tb.Build(0, sync.Map{}, filepath.Join(tempdir, "1.sst"))
	assert.Nil(t, err)
}

func TestBuildSSTTowBlocks(t *testing.T) {
	tb := sst.NewTableBuilder(16)
	tb.Add("11", "11")
	tb.Add("22", "22")
	tb.Add("33", "33")
	tb.Add("44", "44")
	tb.Add("55", "55")
	tb.Add("66", "66")
	assert.Greater(t, tb.Len(), uint32(2))
	tempdir := t.TempDir()
	sstable, err := tb.Build(0, sync.Map{}, filepath.Join(tempdir, "1.sst"))
	assert.Nil(t, err)
	assert.NotNil(t, sstable)
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

func KeyOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("key_%0*d", 8, idx))
}

func ValueOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("value_%0*d", 8, idx))
}

func generateSST(tempdirFn func() string) (*sst.Table, string, error) {

	tb := sst.NewTableBuilder(4096)
	for idx := uint64(0); idx < 1000; idx++ {
		key, value := KeyOf(idx), ValueOf(idx)
		tb.AddByte(key, value)
	}
	tempdir := tempdirFn()
	fp := filepath.Join(tempdir, "1.sst")
	sstable, err := tb.Build(1, sync.Map{}, fp)
	return sstable, fp, err
}

func TestGenerateSST(t *testing.T) {
	st, _, err := generateSST(t.TempDir)
	assert.Nil(t, err)
	assert.Nil(t, st.Close())
}

func TestSSTDecode(t *testing.T) {
	sstable, fp, err := generateSST(t.TempDir)
	assert.Nil(t, err)
	fd, err := os.Open(fp)
	assert.Nil(t, err)
	nsstable, err := sst.OpenTableFromFile(0, sync.Map{}, fd)
	assert.Nil(t, err)
	assert.Equal(t, sstable.Meta(), nsstable.Meta())
	assert.Nil(t, sstable.Close())
	assert.Nil(t, nsstable.Close())
}

func TestSSTIter(t *testing.T) {
	sstable, _, err := generateSST(t.TempDir)
	assert.Nil(t, err)
	defer sstable.Close()
	iter := sst.NewIterAndSeekToFirst(sstable)
	for i := 0; i < 5; i++ {
		for j := uint64(0); j < 100; j++ {
			key := iter.Key()
			value := iter.Value()
			assert.Equalf(t, KeyOf(j), key, "expect key %s, actual key: %s", KeyOf(j), key)
			assert.Equalf(t, ValueOf(j), value, "expect key %s, actual key: %s", ValueOf(j), value)
			iter.Next()
		}
		iter.SeekToFirst()
	}
}

func BenchmarkSSTEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		table, _, _ := generateSST(b.TempDir)
		table.Close()
	}
}

func BenchmarkSSTDecode(b *testing.B) {
	st, fp, _ := generateSST(b.TempDir)
	st.Close()
	for i := 0; i < b.N; i++ {
		fd, _ := os.Open(fp)
		tb, _ := sst.OpenTableFromFile(0, sync.Map{}, fd)
		tb.Close()
	}
}

func BenchmarkSSTIterSeekToFirst(b *testing.B) {
	sstable, _, _ := generateSST(b.TempDir)
	iter := sst.NewIterAndSeekToFirst(sstable)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 100; j++ {
			iter.Next()
		}
	}
	sstable.Close()
}

func BenchmarkSSTIterSeekToKeyExists(b *testing.B) {
	sstable, _, _ := generateSST(b.TempDir)
	iter := sst.NewIterAndSeekToFirst(sstable)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 100; j++ {
			iter.SeekToFirst()
			iter.SeekToKey(KeyOf(j))
		}
	}
	sstable.Close()
}

func BenchmarkSSTIterSeekToKeyNonExists(b *testing.B) {
	sstable, _, _ := generateSST(b.TempDir)
	iter := sst.NewIterAndSeekToFirst(sstable)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 100; j++ {
			iter.SeekToFirst()
			iter.SeekToKey(KeyOf(j + 10086))
		}
	}
	sstable.Close()
}
