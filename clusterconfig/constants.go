package clusterconfig

import (
	"time"
)

// Cluster configuration
const (
	NumDC               int           = 3
	DC0                 string        = "0" // centralized manager
	DC1                 string        = "1"
	DC2                 string        = "2"
	DC3                 string        = "3"
	MaxPartitionSize    float64       = 10
	StorageFilename     string        = "storage_log"
	LogTimeInterval     time.Duration = time.Minute * 10
	ReadThreshold       uint64        = 3
	PopulatingInterval  time.Duration = time.Second * 10
	SyncReplicaInterval time.Duration = time.Second * 10
	LogServiceOn        bool          = false
	PopulateServiceOn   bool          = true
	SyncServiceOn       bool          = true
	PrintServiceOn      bool          = false
	CopyEveryWhereOn    bool          = false
	TCON                bool          = false
	MockTransLatencyON  bool          = false
	readReqFiles        int           = 10
	PopulateInfFactor   int           = 20
	SyncInfFactor       int           = 5
)

// Networking configuration
const (
	SERVER0_IP    string = "0.0.0.0" // Central Manager in California
	SERVER0_PORT1 string = "41001"
	SERVER0_PORT2 string = "41002"
	BANDWIDTH0    int    = 320 // 265 Mbit/s

	SERVER1_IP    string = "0.0.0.0" // DC1 in Ireland
	SERVER1_PORT1 string = "41011"
	SERVER1_PORT2 string = "41012"
	BANDWIDTH1    int    = 280 // Mbit/s

	SERVER2_IP    string = "0.0.0.0" // DC2 in Virginia
	SERVER2_PORT1 string = "41021"
	SERVER2_PORT2 string = "41022"
	BANDWIDTH2    int    = 700 // Mbit/s

	SERVER3_IP    string = "0.0.0.0" // DC3 in North California
	SERVER3_PORT1 string = "41031"
	SERVER3_PORT2 string = "41032"
	BANDWIDTH3    int    = 320 // Mbit/s
)

type ServerIPAddr struct {
	ServerIP    string
	ServerPort1 string
	ServerPort2 string
	Bandwidth   int
}

type ServerIPMap map[string]ServerIPAddr
type CentralManagerIPMap map[string]ServerIPAddr

// When scaled to more servers, need to also add more entries here
func (m *ServerIPMap) CreateIPMap() {
	*m = ServerIPMap{
		//		"0": ServerIPAddr{SERVER0_IP, SERVER0_PORT1, SERVER0_PORT2},
		"1": ServerIPAddr{SERVER1_IP, SERVER1_PORT1, SERVER1_PORT2, BANDWIDTH1},
		"2": ServerIPAddr{SERVER2_IP, SERVER2_PORT1, SERVER2_PORT2, BANDWIDTH2},
		"3": ServerIPAddr{SERVER3_IP, SERVER3_PORT1, SERVER3_PORT2, BANDWIDTH3},
	}
}

// Create IP map for central manager
func (m *CentralManagerIPMap) CreateIPMap() {
	*m = CentralManagerIPMap{
		"0": ServerIPAddr{SERVER0_IP, SERVER0_PORT1, SERVER0_PORT2, BANDWIDTH0},
	}
}
