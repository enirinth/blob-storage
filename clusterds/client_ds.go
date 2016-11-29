// Definition of data structures for clients
// Primarily for simulation
package clusterds

// Write request
type WriteReq struct {
	Content string
	Size    float64
}

// Write response
type WriteResp struct {
	PartitionID string
	BlobID      string
}

// Read request
type ReadReq struct {
	PartitionID string
	BlobID      string
}

// Read response
type ReadResp struct {
	Content string
}
