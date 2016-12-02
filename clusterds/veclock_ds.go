// Defines data strctures in distributed locking
package clusterds

// Request types
// Copy a partition to other DCs or delete a partition
const (
	PAR_COPY int = 1
	PAR_DEL  int = 2
)

// Message struct
// A DC pings others with its request type, DC ID and  number of local/global
// Reads at the moment it sends it
type SenderMsg struct {
	SenderID       string
	SenderLocRead  uint64
	SenderGlobRead uint64
	ReqType        int
}

// Request struct
type DCReq struct {
	RequesterID string
	ReqType     int
	ReadStamp   uint64 // Number of global read requests (vector clock)
}

// The request queue stores in each DC
type ReqQueue map[string]*DCReq
