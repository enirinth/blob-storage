package loclock

import (
	ds "github.com/enirinth/blob-storage/clusterds"
	"sync"
)

type StorageLockMap map[string]*sync.RWMutex

// Constructor (according to Storage Map
func (s StorageLockMap) CreateLockMap(storageMap *map[string]*ds.Partition) {
	for k, _ := range *storageMap {
		var m sync.RWMutex
		s[k] = &m
	}
}

func (s StorageLockMap) AddEntry(newPartitionID string) {
	var m sync.RWMutex
	s[newPartitionID] = &m
}

// Reader lock
func (s StorageLockMap) RLock(partitionID string) {
	s[partitionID].RLock()
}

// Reader unlock
func (s StorageLockMap) RUnlock(partitionID string) {
	s[partitionID].RUnlock()
}

// Writer lock
func (s StorageLockMap) WLock(partitionID string) {
	s[partitionID].Lock()
}

// Writer unlock
func (s StorageLockMap) WUnlock(partitionID string) {
	s[partitionID].Unlock()
}
