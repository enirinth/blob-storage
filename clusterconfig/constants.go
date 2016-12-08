package clusterconfig

import (
	"time"
)

// Cluster configuration
const (
	NumDC               int           = 3
	DC0   		        string        = "0"      // centralized manager
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
	SERVER0_IP      string = "0.0.0.0"
	SERVER0_PORT1   string = "41001"
	SERVER0_PORT2   string = "41002"

	SERVER1_IP    string = "52.209.171.220"
	SERVER1_PORT1 string = "41011"
	SERVER1_PORT2 string = "41012"

	SERVER2_IP    string = "54.221.133.142"
	SERVER2_PORT1 string = "41021"
	SERVER2_PORT2 string = "41022"

	SERVER3_IP    string = "54.153.39.155"
	SERVER3_PORT1 string = "41031"
	SERVER3_PORT2 string = "41032"
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
		"0": ServerIPAddr{SERVER0_IP, SERVER0_PORT1, SERVER0_PORT2},
		"1": ServerIPAddr{SERVER1_IP, SERVER1_PORT1, SERVER1_PORT2},
		"2": ServerIPAddr{SERVER2_IP, SERVER2_PORT1, SERVER2_PORT2},
		"3": ServerIPAddr{SERVER3_IP, SERVER3_PORT1, SERVER3_PORT2},
	}
}
