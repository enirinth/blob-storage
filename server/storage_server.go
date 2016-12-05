package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"github.com/enirinth/blob-storage/locking/loclock"
	"github.com/enirinth/blob-storage/util"
	"net"
	"net/http"
	"net/rpc"
	"os"
	//"sync"
	"time"
)

const (
	numDC               int           = config.NumDC            // total number of DCs
	MaxPartitionSize    float64       = config.MaxPartitionSize // maximum size of partition (excluding metadata)
	storage_log         string        = config.StorageFilename  // file path that logs the storage
	logInterval         time.Duration = config.LogTimeInterval  // time interval to log storage
	populatingInterval  time.Duration = time.Second * 10        // time interval to scan partitions and decide whether to populate the partition to other datacenter
	syncReplicaInterval time.Duration = time.Second * 10        //time interval to synchronize replica across DCs
	readThreshold       uint64        = 3                       // Threshold number for read to trigger populating
)

type (
	Listener int
)

var (
	DCID string // ID of current DC

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
// This will automatically create a handling thread
func (l *Listener) HandleWriteReq(req ds.WriteReq, resp *ds.WriteResp) error {
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

	return nil
}

// Read request handler
// This will automatically create a handling thread
func (l *Listener) HandleReadReq(req ds.ReadReq, resp *ds.ReadResp) error {
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

	return nil
}

// Populate one partition to all DCs
// Periodically scan read-count map and make decisiions independently
func populateReplica() {
	t := time.NewTicker(populatingInterval)
	for {
		// Scan read map
		// TODO: add fine grained locking here
		fmt.Println("Start populating data")
		for partitionID, count := range ReadMap {
			if count.LocalRead >= readThreshold {
				// Send partition to all other  DC
				for dcID, ipAddr := range IPMap {
					// Only send to DC that haven't got this partition/replica
					if dcID != DCID && !util.FindDC(dcID, ReplicaMap[partitionID]) {
						go func(serverIP string, serverPort string, pID string) {
							client, err := rpc.DialHTTP("tcp", serverIP+":"+serverPort)
							if err != nil {
								fmt.Println("Dial HTTP error in populating replica. ")
								log.Fatal(err)
							}
							var msg = *storageTable[pID]
							var reply bool
							err = client.Call("Listener.HandleIncomingReplica", msg, &reply)
							if err != nil {
								log.Fatal(err)
							}
							// Update replica map after sending partition
							ReplicaMap[partitionID].AddDC(dcID)
							// No need to update readmap or storage table, because already have that partition (this is the sender itself)
						}(ipAddr.ServerIP, ipAddr.ServerPort1, partitionID)
					}
				}
			}
		}

		// One scan finishes, signals time ticker
		<-t.C
	}

}

// Add partition to local storageTable upon receiving another DC's populating request
// TODO: add locking here
func (l *Listener) HandleIncomingReplica(newPartition *ds.Partition, resp *bool) error {
	// Add new partition to storage table
	storageTable[newPartition.PartitionID] = newPartition
	// Update replica map after receiving new parition
	if _, ok := ReplicaMap[newPartition.PartitionID]; !ok {
		// New partition not in replica map, create new entry
		ReplicaMap[newPartition.PartitionID] = &ds.PartitionState{newPartition.PartitionID, []string{DCID}}
	} else {
		// New partition already in replica map, just add self-DCID to DCList
		ReplicaMap[newPartition.PartitionID].AddDC(DCID)
	}
	// Add new entry in read map after receiving new partition
	ReadMap[newPartition.PartitionID] = &ds.NumRead{0, 0}
	// Print storage table after adding new partition
	fmt.Println("Storage Table after receiving replica:")
	util.PrintStorage(&storageTable)
	fmt.Println("------")

	*resp = true
	return nil
}

// Periodic service that synchronizes paritions with the same paritionID
// i.e. a replica set
// TODO: fine-grained locking
func syncReplica() {
	t := time.NewTicker(syncReplicaInterval)
	for {
		// Scan replica map
		for partitionID, partitionState := range ReplicaMap {
			for _, id := range (*partitionState).DCList {
				// Sync current partition with all OTHER DC(s) that store it
				if DCID != id {
					go func(pID string, dcID string) {
						client, err := rpc.DialHTTP("tcp", IPMap[dcID].ServerIP+":"+IPMap[dcID].ServerPort1)
						if err != nil {
							fmt.Println("Dial HTTP error in sync replica.")
							log.Fatal(err)
						}
						var msg = *storageTable[partitionID]
						var reply ds.Partition
						err = client.Call("Listener.HandleSyncReplica", msg, &reply)
						if err != nil {
							log.Fatal(err)
						}
						// Merge reply of other DCs to local partition
						util.MergePartition(storageTable[partitionID], &reply)
					}(partitionID, id)
				}
			}
		}

		<-t.C
	}
}

// Handler that merge a incoming partiion to local replication
// TODO: locking
func (l *Listener) HandleSyncReplica(comingPartition *ds.Partition, resp *ds.Partition) error {
	// Store current partition to reply
	*resp = *storageTable[comingPartition.PartitionID]
	// Merge current partition and incoming partition
	util.MergePartition(comingPartition, storageTable[comingPartition.PartitionID])
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
	// Parse DCID from command line
	switch id := os.Args[1]; id {
	case "1":
		DCID = config.DC1
	case "2":
		DCID = config.DC2
	case "3":
		DCID = config.DC3
	default:
		log.Fatal(errors.New("Error parsing DCID from command line"))
	}

	fmt.Println("Storage server starts")

	// Initiates a thread that periodically persist storage into a log file (on disk)
	go persistStorage(&storageTable)

	// Initiates populating service
	go populateReplica()

	// Initiaes replica synchronization service
	//go syncReplica()

	// Main loop
	listener := new(Listener)
	rpc.Register(listener)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+IPMap[DCID].ServerPort1)
	if e != nil {
		log.Fatal(e)
	}
	http.Serve(l, nil)
}
