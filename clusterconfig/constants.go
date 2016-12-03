package clusterconfig

import (
	"time"
)

// Cluster configuration
const (
	NumDC            int           = 3
	MaxPartitionSize float64       = 10
	StorageFilename  string        = "storage_log"
	LogTimeInterval  time.Duration = time.Minute * 10
)

// Networking configuration
const (
	SERVER_IP1   string = "0.0.0.0"
	SERVER_Port1 string = "42001"

	SERVER_IP2   string = "0.0.0.0"
	SERVER_Port2 string = "42002"

	SERVER_IP3   string = "0.0.0.0"
	SERVER_Port3 string = "42003"
)
