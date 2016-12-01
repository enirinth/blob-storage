package clusterLocking

import (
	ds "github.com/enirinth/blob-storage/clusterds"
	"sync"
)

var StorageLockMap map[string]*sync.RWMutex

// Constructor (according to Storage Map
func (*StorageLockMap) createLockMap(storageMap *map[string]*ds.Partition) {
	for k, _ := range storageMap {
		var m sync.RWMutex
		StorageLockMap[k] = &m
	}
}

func (*StorageLockMap) addEntry(newPartitionID string) {
	var m sync.RWMutex
	StorageLockMap[newPartitionID] = &m
}

// Reader lock
func (*StorageLockMap) RLock(partitionID string) {
	StorageMap[partitionID].RLock()
}

// Reader unlock
func (*StorageLockMap) RUnlock(partitionID string) {
	StorageMap[partitionID].RUnlock()
}

// Writer lock
func (*StorageLockMap) WLock(partitionID string) {
	StorageMap[partitionID].Lock()
}

// Writer unlock
func (*StorageLockMap) WUnlock(partitionID string) {
	StorageMap[partitionID].Unlock()
}
