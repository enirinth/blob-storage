// Locking for StorageTable in cluster manager
// There is a detailed explannation on why there is no need for rw mutex here:
// https://docs.google.com/presentation/d/1cjpXEeILcemgNt2MFYXDgC8MdpLFgimEDZpN9zooHck/edit#slide=id.p
// Go doesn't have a built-in atomic map, therefore it needs to be handled manually
package loclock

import (
	ds "github.com/enirinth/blob-storage/clusterds"
	"sync"
)

type StorageTableLockMap map[string]*sync.Mutex

// Constructor (according to Storage Map
// Called upon initializing of cluster manager
func (s *StorageTableLockMap) CreateLockMap(storageTable *map[string]*ds.Partition) {
	*s = make(map[string]*sync.Mutex) // Construct lock map, otherwise it's nil
	for partitionID, _ := range *storageTable {
		(*s)[partitionID] = new(sync.Mutex)
	}
}

// Add a new entry in lock map
// Could be used when creating a new partition
func (s StorageTableLockMap) AddEntry(newPartitionID string) {
	s[newPartitionID] = new(sync.Mutex)
}

// (Write) lock
func (s StorageTableLockMap) Lock(partitionID string) {
	if _, ok := s[partitionID]; !ok {
		handleError(partitionID)
	}
	s[partitionID].Lock()
}

// (Writer) unlock
func (s StorageTableLockMap) Unlock(partitionID string) {
	if _, ok := s[partitionID]; !ok {
		handleError(partitionID)
	}
	s[partitionID].Unlock()
}
