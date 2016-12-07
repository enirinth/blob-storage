package clusterconfig

import (
	"time"
)

// Cluster configuration
const (
	NumDC               int           = 3
	DC1                 string        = "1"
	DC2                 string        = "2"
	DC3                 string        = "3"
	MaxPartitionSize    float64       = 10
	StorageFilename     string        = "storage_log"
	LogTimeInterval     time.Duration = time.Minute * 10
	ReadThreshold       uint64        = 3
	PopulatingInterval  time.Duration = time.Second * 10
	SyncReplicaInterval time.Duration = time.Second * 10
)

// Networking configuration
const (
	SERVER1_IP    string = "0.0.0.0"
	SERVER1_PORT1 string = "42011"
	SERVER1_PORT2 string = "42012"

	SERVER2_IP    string = "0.0.0.0"
	SERVER2_PORT1 string = "42021"
	SERVER2_PORT2 string = "42022"

	SERVER3_IP    string = "0.0.0.0"
	SERVER3_PORT1 string = "42031"
	SERVER3_PORT2 string = "42032"
)

type ServerIPAddr struct {
	ServerIP    string
	ServerPort1 string
	ServerPort2 string
}

type ServerIPMap map[string]ServerIPAddr

// When scaled to more servers, need to also add more entries here
func (m *ServerIPMap) CreateIPMap() {
	*m = ServerIPMap{
		"1": ServerIPAddr{SERVER1_IP, SERVER1_PORT1, SERVER1_PORT2},
		"2": ServerIPAddr{SERVER2_IP, SERVER2_PORT1, SERVER2_PORT2},
		"3": ServerIPAddr{SERVER3_IP, SERVER3_PORT1, SERVER3_PORT2},
	}
}
