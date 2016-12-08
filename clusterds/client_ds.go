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
	Size    float64
}

// Read response from central manager, return DC location for that blob
type CentralManagerReadResp struct {
	Address     string
	Size        float64
}

type CentralDCReadReq struct {
	PartitionID string
	BlobID      string
	Size        float64
}

type CentralDCReadResp struct {
	Content  string
}