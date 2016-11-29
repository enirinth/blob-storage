package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	ds "github.com/enirinth/read-clock/clusterds"
	util "github.com/enirinth/read-clock/util"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const maxPartitionSize = 10 // maximum size of partition (excluding metadata)
const numDC int = 3         // total number of DCs

type Listener int

// Cluster map data structures
var ReplicaMap = make(map[string]ds.PartitionState)

var ReadMap = make(map[string]ds.NumRead)

// Local (intra-DC) storage data structures
var storageTable = make(map[string]*ds.Partition)

// Persist storage into a log file
func persistStorage(table *map[string]*ds.Partition) {
	time.Sleep(time.Second * 10)
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
}

// Write request handler
// TODO: deal with replica map and read map
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

		// Select the first partition that is not full (different from Ambry)
		// TODO: Need fine-grained locking here
		partitionID := ""
		for id, partition := range storageTable {
			if partition.PartitionSize+size <= maxPartitionSize {
				partitionID = id
				break
			}
		}

		if len(partitionID) == 0 { // If all partitions are full create a new one
			partitionID, err = util.NewUUID()
			if err != nil {
				log.Fatal(err)
			}
			storageTable[partitionID] = &(ds.Partition{partitionID, []ds.Blob{{blobUUID, content, size, now}}, now, size})
		} else {
			// Add blob to storageTable
			storageTable[partitionID].AppendBlob(ds.Blob{blobUUID, content, size, now})
			storageTable[partitionID].PartitionSize += size
		}

		// Reply with (PartitionID, blobID) pair
		*resp = ds.WriteResp{"1", blobUUID}

		// Print storage table after write
		fmt.Println("Storage Table after update:")
		util.PrintStorage(&storageTable)
		fmt.Println("------")
	}(req, resp)

	wg.Wait() // wait for handler thread to finish, then return reply message
	return nil
}

func init() {
	// Log settings
	log.SetFormatter(&log.JSONFormatter{})
	f, err := os.OpenFile("storage_log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	// Initiates a thread that periodically persist storage into a log file (on disk)
	go persistStorage(&storageTable)
}

// Server main loop
func main() {
	fmt.Println("Storage server starts")

	// Main loop
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
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
