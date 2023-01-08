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
	tb.Build(0, sync.Map{}, filepath.Join(tempdir, "1.sst"))
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
	return s2b(fmt.Sprintf("key_%d", idx))
}

func ValueOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("value_%d", idx))
}

func generateSST(getTempDir func() string) (*sst.Table, string) {
	tb := sst.NewTableBuilder(128)
	for idx := uint64(0); idx < 100; idx++ {
		key, value := KeyOf(idx), ValueOf(idx)
		tb.AddByte(key, value)
	}
	tempdir := getTempDir()
	fp := filepath.Join(tempdir, "1.sst")
	sstable, err := tb.Build(0, sync.Map{}, fp)
	if err != nil {
		panic(err)
	}
	return sstable, fp
}

func TestGenerateSST(t *testing.T) {
	generateSST(t.TempDir)
}

func TestSSTDecode(t *testing.T) {
	sstable, fp := generateSST(t.TempDir)
	fd, err := os.Open(fp)
	assert.Nil(t, err)
	nsstable, err := sst.OpenTableFromFile(0, sync.Map{}, fd)
	assert.Nil(t, err)
	assert.Equal(t, sstable.Meta(), nsstable.Meta())
}

func TestSSTIter(t *testing.T) {
	sstable, _ := generateSST(t.TempDir)
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
		generateSST(b.TempDir)
	}
}

func BenchmarkSSTDecode(b *testing.B) {
	_, fp := generateSST(b.TempDir)
	for i := 0; i < b.N; i++ {
		fd, _ := os.Open(fp)
		sst.OpenTableFromFile(0, sync.Map{}, fd)
	}
}

func BenchmarkSSTIter(b *testing.B) {
	sstable, _ := generateSST(b.TempDir)
	iter := sst.NewIterAndSeekToFirst(sstable)
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 100; j++ {
			iter.Next()
		}
	}
}
