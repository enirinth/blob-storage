package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"github.com/enirinth/blob-storage/locking/loclock"
	"github.com/enirinth/blob-storage/util"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	DCID             string        = config.DC1              // Id of current DC
	numDC            int           = config.NumDC            // total number of DCs
	MaxPartitionSize float64       = config.MaxPartitionSize // maximum size of partition (excluding metadata)
	storage_log      string        = config.StorageFilename  // file path that logs the storage
	logInterval      time.Duration = config.LogTimeInterval  // time interval to log storage
	repServInterval  time.Duration = time.Second * 10        // time interval to scan partitions and decide whether to populate the partition to other datacenter
	readThreshold    uint64        = 5                       // Threshold number for read to trigger populating
)

type (
	Listener int
)

var (
	// Routing
	IPMap config.ServerIPMap
	// Cluster map data structures
	ReplicaMap = make(map[string]*ds.PartitionState)

	// Local (intra-DC) storage data structures
	storageTable = make(map[string]*ds.Partition)
	ReadMap      = make(map[string]*ds.NumRead)

	// Locking
	rcLock loclock.ReadCountLockMap // Fined-grained locking for ReadMap (read-count map)
)

// Persist storage into a log file
// TODO: add locking
func persistStorage(table *map[string]*ds.Partition) {
	// Periodically log storage
	t := time.NewTicker(logInterval)
	for {
		fmt.Println("Starting to persist storage")
		// Delete old storage log
		_, e := os.OpenFile(storage_log, os.O_WRONLY|os.O_CREATE, 0755) // if file exists
		if e == nil {
			err := os.Remove(storage_log)
			if err != nil {
				fmt.Println(err.Error())
				log.Fatal(err)
			}
		}
		// Log latest storage
		for partitionID, partition := range *table {
			// Mark partition beginning
			log.WithFields(log.Fields{
				"Partition starts": "-----------------------------",
			}).Info("Partition " + partitionID)

			// Persist partition information
			log.WithFields(log.Fields{
				"CreateTimestamp": partition.CreateTimestamp,
				"PartitionSize":   partition.PartitionSize,
			}).Info("Partition " + partitionID)

			// Persist all blobs from a certain partition
			for _, blob := range (*partition).BlobList {
				log.WithFields(log.Fields{
					"BlobID":          blob.BlobID,
					"Content":         blob.Content,
					"BlobSize":        blob.BlobSize,
					"CreateTimestamp": blob.CreateTimestamp,
				}).Info("Blob " + blob.BlobID)
			}

			// Mark partition ending
			log.WithFields(log.Fields{
				"Partition ends": "-------------------------------",
			}).Info("Partition " + partitionID)
		}
		<-t.C
	}
}

// Write request handler
func (l *Listener) HandleWriteReq(req ds.WriteReq, resp *ds.WriteResp) error {
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a new thread handling request
	go func(req ds.WriteReq, resp *ds.WriteResp) {
		defer wg.Done()

		// Parse write request, get blob info
		content := req.Content
		size := req.Size
		now := time.Now().Unix()
		blobUUID, err := util.NewUUID()
		if err != nil {
			log.Fatal(err)
		}

		// Select the first partition that is not full (this is different from Ambry)
		partitionID := ""
		for id, partition := range storageTable {
			if partition.PartitionSize+size <= MaxPartitionSize {
				partitionID = id
				break
			}
		}

		if len(partitionID) == 0 {
			// If all partitions are full create a new one
			partitionID, err = util.NewUUID()
			if err != nil {
				log.Fatal(err)
			}
			// Create new entries in storage table
			storageTable[partitionID] = &(ds.Partition{partitionID, []ds.Blob{{blobUUID, content, size, now}}, now, size})
			// Crete new entries in replica map and read map
			ReadMap[partitionID] = &(ds.NumRead{0, 0})
			ReplicaMap[partitionID] = &(ds.PartitionState{partitionID, []string{DCID}})
			// Create new entries in lock map
			rcLock.AddEntry(partitionID)
		} else {
			// Add blob to storageTable
			storageTable[partitionID].AppendBlob(ds.Blob{blobUUID, content, size, now})
			storageTable[partitionID].PartitionSize += size
		}

		// Reply with (PartitionID, blobID) pair
		*resp = ds.WriteResp{partitionID, blobUUID}

		// Print storage table after write
		fmt.Println("Storage Table after update:")
		util.PrintStorage(&storageTable)
		fmt.Println("------")
	}(req, resp)

	wg.Wait() // wait for handler thread to finish, then return reply message
	return nil
}

// Read request handler
func (l *Listener) HandleReadReq(req ds.ReadReq, resp *ds.ReadResp) error {
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a new thread handling request
	go func(req ds.ReadReq, resp *ds.ReadResp) {
		defer wg.Done()

		// Parse read request
		partitionID := req.PartitionID
		blobID := req.BlobID

		// Look for target blob
		if _, ok := storageTable[partitionID]; ok {
			for _, blob := range storageTable[partitionID].BlobList {
				if blob.BlobID == blobID {
					*resp = ds.ReadResp{blob.Content, blob.BlobSize}
					break
				}
			}
		}

		if resp.Size != 0 {
			// Update read count
			rcLock.WLock(partitionID)
			// ReadMap[partitionID].GlobalRead += 1
			ReadMap[partitionID].LocalRead += 1
			rcLock.WUnlock(partitionID)
		} else {
			// Not found
			*resp = ds.ReadResp{"NOT_FOUND", 0}
		}
	}(req, resp)

	wg.Wait()
	return nil
}

// Populate one partition to all DCs
// Periodically scan read-count map and make decisiions independently
func sendReplica() {
	t := time.NewTicker(repServInterval)
	for {
		// Scan read map
		// TODO: add fine grained locking here
		for partitionID, count := range ReadMap {
			if count.LocalRead >= readThreshold {
				// Send partition to all other  DC
				for dcID, ipAddr := range IPMap {
					// Only send to DC that haven't got this partition/replica
					if dcID != DCID && !util.FindDC(dcID, ReplicaMap[partitionID]) {
						go func(serverIP string, serverPort string, pID string) {
							client, err := rpc.Dial("tcp", serverIP+":"+serverPort)
							if err != nil {
								log.Fatal(err)
							}
							var msg = *storageTable[pID]
							var reply bool
							err = client.Call("Listener.HandleIncomingReplica", msg, &reply)
							if err != nil {
								log.Fatal(err)
							}
						}(ipAddr.ServerIP, ipAddr.ServerPort, partitionID)
					}
				}
			}
		}

		// One scan finishes, signals time ticker
		<-t.C
	}

}

// Add partition to local storageTable upon receiving another DC's populating request
func (l *Listener) HandleIncomingReplica(newPartition *ds.Partition, resp *bool) error {
	storageTable[newPartition.PartitionID] = newPartition
	*resp = true
	return nil
}

func init() {
	// Log settings
	log.SetFormatter(&log.JSONFormatter{})
	_, e := os.OpenFile(storage_log, os.O_WRONLY|os.O_CREATE, 0755)
	if e != nil {
		// If log file doesn't exist, create new one
		_, err := os.Create(storage_log)
		if err != nil {
			fmt.Println(err.Error())
			log.Fatal(err)
		}
	}
	f, err := os.OpenFile(storage_log, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	// Setup locking
	rcLock.CreateLockMap(&ReadMap)
	// Setup routing
	IPMap.CreateIPMap()
}

// Server main loop
func main() {
	fmt.Println("Storage server starts")

	// Initiates a thread that periodically persist storage into a log file (on disk)
	go persistStorage(&storageTable)

	// Initiates populating service
	//go sendReplica()

	// Main loop
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+config.SERVER1_PORT1)
	if err != nil {
		log.Fatal(err)
	}
	inbound, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)
}
