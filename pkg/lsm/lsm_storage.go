package lsm

import (
	"fmt"
	"sync"

	"github.com/dashjay/mini-lsm-go/pkg/iterator"
	"github.com/dashjay/mini-lsm-go/pkg/memtable"
	"github.com/dashjay/mini-lsm-go/pkg/sst"
)

type StorageInner struct {
	memt       *memtable.Table
	immMemt    []*memtable.Table
	l0SSTables []*sst.Table
	levels     [][]*sst.Table
	nextSSTID  uint32
}

func NewStorageInner() *StorageInner {
	return &StorageInner{
		memt:       memtable.NewTable(),
		immMemt:    make([]*memtable.Table, 0),
		l0SSTables: make([]*sst.Table, 0),
		levels:     make([][]*sst.Table, 0),
		nextSSTID:  1,
	}
}

type Storage struct {
	inner      *StorageInner
	mu         sync.RWMutex
	flushLock  sync.Mutex
	path       string
	blockCache sync.Map
}

func NewStorage(path string) *Storage {
	return &Storage{
		inner:      NewStorageInner(),
		mu:         sync.RWMutex{},
		flushLock:  sync.Mutex{},
		path:       path,
		blockCache: sync.Map{},
	}
}

func (s *Storage) Get(key []byte) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val := s.inner.memt.Get(key)
	if val != nil {
		return val
	}
	for _, mt := range s.inner.immMemt {
		if val := mt.Get(key); val != nil {
			return val
		}
	}

	iters := make([]iterator.Iter, 0, len(s.inner.l0SSTables))
	for t := range s.inner.l0SSTables {
		iters = append(iters, sst.NewIterAndSeekToKey(s.inner.l0SSTables[t], key))
	}
	iter := iterator.NewmergeIterator(iters...)
	if iter.IsValid() {
		return iter.Key()
	}
	return nil
}

func (s *Storage) Put(key, value []byte) {
	if len(value) == 0 {
		panic("value cannot be empty")
	}
	if len(key) == 0 {
		panic("key cannot be empty")
	}
	s.mu.RLock()
	defer s.mu.RLock()
	s.inner.memt.Put(key, value)
}

func (s *Storage) Delete(key []byte) {
	if len(key) == 0 {
		panic("key cannot be empty")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.inner.memt.Put(key, nil)
}

func (s *Storage) sstPath(id uint32) string {
	return fmt.Sprintf("%d.sst", id)
}

func (s *Storage) Sync() error {
	s.flushLock.Lock()
	defer s.flushLock.Unlock()

	// 1. mark mmtable as imm_mmtable
	s.mu.Lock()
	newMemtable := memtable.NewTable()

	oldMemtable := s.inner.memt
	flushMemtable := oldMemtable

	// 2. replace it with a new memtable
	s.inner.memt = newMemtable

	sstId := s.inner.nextSSTID

	// 3. append memtable to immemtable
	s.inner.immMemt = append(s.inner.immMemt, oldMemtable)
	s.mu.Unlock()

	// 4. flush memtable to a new table builder
	builder := sst.NewTableBuilder(4096)
	flushMemtable.Flush(builder)

	sst, err := builder.Build(sstId, s.blockCache, s.sstPath(sstId))
	if err != nil {
		return err
	}

	s.inner.immMemt = s.inner.immMemt[:len(s.inner.immMemt)-1]
	s.inner.l0SSTables = append(s.inner.l0SSTables, sst)
	s.inner.nextSSTID += 1
	return nil
}

func (s *Storage) Scan(lower, upper []byte) {

}
