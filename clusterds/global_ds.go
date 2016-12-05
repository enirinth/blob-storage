// Definition of data structures for cluster manager
package clusterds

// ClusterMap -- mapping partition to a list of DCs that store it
type PartitionState struct {
	PartitionID string
	DCList      []string
}

func (ps *PartitionState) AddDC(dcID string) {
	(*ps).DCList = append((*ps).DCList, dcID)
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
