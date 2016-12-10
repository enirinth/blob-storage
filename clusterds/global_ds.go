// Definition of data structures for cluster manager
package clusterds

// ClusterMap -- mapping partition to a list of DCs that store it
type PartitionState struct {
	PartitionID string
	DCList      []string
}

// Add DC to a partitionState (store list of a certain partition)
// In set semantics (no duplicates)
func (ps *PartitionState) AddDC(dcID string) {
	if !ps.FindDC(dcID) {
		(*ps).DCList = append((*ps).DCList, dcID)
	}
}

// Find if a certain DC stores a certain partition
func (ps *PartitionState) FindDC(dcID string) bool {
	for _, id := range (*ps).DCList {
		if id == dcID {
			return true
		}
	}
	return false
}

// Store two types of read counts
type NumRead struct {
	LocalRead  uint64
	GlobalRead uint64
}

// Actual storage
type Partition struct {
	PartitionID     string
	BlobList        []Blob
	CreateTimestamp int64
	PartitionSize   float64
}

// Append blob to a partition ('s blob list)
func (p *Partition) AppendBlob(b Blob) {
	(*p).BlobList = append((*p).BlobList, b)
}

// Fake blob object
type Blob struct {
	BlobID          string
	Content         string
	BlobSize        float64
	CreateTimestamp int64
}
