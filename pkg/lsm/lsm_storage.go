package lsm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	iter := iterator.NewMergeIterator(iters...)
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
	defer s.mu.RUnlock()
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
	return filepath.Join(s.path, fmt.Sprintf("%d.sst", id))
}

func (s *Storage) MakeNewMemtable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inner.memt, s.inner.immMemt = memtable.NewTable(), append(s.inner.immMemt, s.inner.memt)
}

func (s *Storage) SinkImemtableToSST() error {
	s.flushLock.Lock()
	defer s.flushLock.Unlock()
	sstId := s.inner.nextSSTID
	s.mu.Lock()
	defer s.mu.Unlock()
	flushMemtable := s.inner.immMemt[len(s.inner.immMemt)-1]
	builder := sst.NewTableBuilder(4096)
	flushMemtable.Flush(builder)

	sstable, err := builder.Build(sstId, s.blockCache, s.sstPath(sstId))
	if err != nil {
		return err
	}

	s.inner.immMemt = s.inner.immMemt[:len(s.inner.immMemt)-1]
	s.inner.l0SSTables = append([]*sst.Table{sstable}, s.inner.l0SSTables...)
	s.inner.nextSSTID += 1
	return nil
}

func (s *Storage) DebugScan(lower, upper []byte) iterator.Iter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var iters []iterator.Iter
	memtScan := s.inner.memt.Scan(lower, upper)
	iters = append(iters, memtScan)

	for _, mt := range s.inner.immMemt {
		imemtScan := mt.Scan(lower, upper)
		iters = append(iters, imemtScan)
	}

	for t := range s.inner.l0SSTables {
		sstIter := sst.NewIterAndSeekToKey(s.inner.l0SSTables[t], lower)
		iters = append(iters, sstIter)
	}

	return iterator.NewMergeIterator(iters...)
}

func (s *Storage) Scan(lower, upper []byte) iterator.Iter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var iters []iterator.Iter

	iters = append(iters, s.inner.memt.Scan(lower, upper))

	for _, mt := range s.inner.immMemt {
		iters = append(iters, mt.Scan(lower, upper))
	}

	for t := range s.inner.l0SSTables {
		iters = append(iters, sst.NewIterAndSeekToKey(s.inner.l0SSTables[t], lower))
	}
	return iterator.NewMergeIterator(iters...)
}

func (s *Storage) Compact() {
	log.Printf("compact with l0SSTables: %d", len(s.inner.l0SSTables))
	if len(s.inner.l0SSTables) >= 2 {
		s.mu.RLock()
		l0SSTableLength := len(s.inner.l0SSTables)
		sn := s.inner.l0SSTables[l0SSTableLength-1]
		snID := sn.SSTID()
		snm1 := s.inner.l0SSTables[l0SSTableLength-2]
		snm1ID := snm1.SSTID()
		s.mu.RUnlock()

		snIter := sst.NewIterAndSeekToFirst(sn)
		snm1Iter := sst.NewIterAndSeekToFirst(snm1)
		mergeIter := iterator.NewTwoMerger(snm1Iter, snIter)
		builder := sst.NewTableBuilder(4096)
		for mergeIter.IsValid() {
			builder.AddByte(mergeIter.Key(), mergeIter.Value())
			mergeIter.Next()
		}
		sstId := s.inner.nextSSTID
		sstable, err := builder.Build(sstId, s.blockCache, s.sstPath(sstId))
		if err != nil {
			log.Printf("sstable build fail: %s", err)
		}
		s.mu.Lock()
		defer func() {
			snm1.Close()
			sn.Close()
			os.Remove(s.sstPath(snID))
			os.Remove(s.sstPath(snm1ID))
		}()
		if s.inner.l0SSTables[l0SSTableLength-1].SSTID() == snID &&
			s.inner.l0SSTables[l0SSTableLength-2].SSTID() == snm1ID {
			for _, sst := range s.inner.l0SSTables {
				log.Printf("sst: %d\n", sst.SSTID())
			}
			log.Println()
			s.inner.l0SSTables = append(s.inner.l0SSTables[:l0SSTableLength-2], sstable)
			for _, sst := range s.inner.l0SSTables {
				log.Printf("sst: %d\n", sst.SSTID())
			}
			log.Println()
		}
		s.mu.Unlock()
	}

}
