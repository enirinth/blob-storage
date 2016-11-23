// Definition of data structures for clients
// Primarily for simulation
package clusterds

// Write message
type WriteMsg struct {
	Content string
	Size    float64
}

// Read message
type ReadMsg struct {
	PartitionID string
	BlobID      string
}
