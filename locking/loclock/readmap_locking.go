// Locking for ReadMap in cluster manager
// Every read request will increase the accumulated read count for its target partition
// Go doesn't have a built-in atomic map, therefore it needs to be handled manually
package loclock

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	ds "github.com/enirinth/blob-storage/clusterds"
	"sync"
)

type ReadCountLockMap map[string]*sync.RWMutex

// Error when trying to lock a partition that doesn't exist
func handleRCError(partitionID string) {
	err := errors.New("PartitionID: " + partitionID +
		" does not exist in readcount lock map, cannot lock")
	fmt.Println(err.Error())
	log.Fatal(err)
}

// Constructor (according to Storage Map
// Called when cluster manager initializing
func (s *ReadCountLockMap) CreateLockMap(readMap *map[string]*ds.NumRead) {
	*s = make(map[string]*sync.RWMutex) // Construct lock map, otherwise it's nil
	for partitionID, _ := range *readMap {
		(*s)[partitionID] = new(sync.RWMutex)
	}
}

// Add a new entry in lock map
// Could be used when creating a new partition
func (s ReadCountLockMap) AddEntry(newPartitionID string) {
	s[newPartitionID] = new(sync.RWMutex)
}

// Reader lock
func (s ReadCountLockMap) RLock(partitionID string) {
	if _, ok := s[partitionID]; !ok {
		handleRCError(partitionID)
	}
	s[partitionID].RLock()
}

// Reader unlock
func (s ReadCountLockMap) RUnlock(partitionID string) {
	if _, ok := s[partitionID]; !ok {
		handleRCError(partitionID)
	}
	s[partitionID].RUnlock()
}

// Writer lock
func (s ReadCountLockMap) WLock(partitionID string) {
	if _, ok := s[partitionID]; !ok {
		handleRCError(partitionID)
	}
	s[partitionID].Lock()
}

// Writer unlock
func (s ReadCountLockMap) WUnlock(partitionID string) {
	if _, ok := s[partitionID]; !ok {
		handleRCError(partitionID)
	}
	s[partitionID].Unlock()
}
