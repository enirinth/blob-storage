// Definition of data structures for cluster manager
package clusterds

// ClusterMap
type PartitionState struct {
	PartitionID string
	DCList      []string
}

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

func (Par *Partition) AppendBlob(B Blob) {
	Par.BlobList = append(Par.BlobList, B)
}

type Blob struct {
	BlobID          string
	Content         string
	BlobSize        float64
	CreateTimestamp int64
}
