package test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"unsafe"

	"github.com/dashjay/mini-lsm-go/pkg/sst"
)

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	// // nolint:govet // unsafe for transfer string to []byte
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func BigKeyOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("big_key_%0*d", 16, idx))
}

func BigValueOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("big_value_%0*d", 16, idx))
}

func KeyOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("key_%0*d", 8, idx))
}

func ValueOf(idx uint64) []byte {
	return s2b(fmt.Sprintf("value_%0*d", 8, idx))
}

const (
	GenerateBlockSize = 4096
)

func GenerateSST(tempdirFn func() string, keyCount uint64) (*sst.Table, string, error) {
	tb := sst.NewTableBuilder(GenerateBlockSize)
	for idx := uint64(0); idx < keyCount; idx++ {
		key, value := KeyOf(idx), ValueOf(idx)
		tb.AddByte(key, value)
	}
	tempdir := tempdirFn()
	fp := filepath.Join(tempdir, "1.sst")
	sstable, err := tb.Build(1, &sync.Map{}, fp)
	return sstable, fp, err
}
