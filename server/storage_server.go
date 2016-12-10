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
	"time"
)

const (
	// total number of DCs
	numDC int = config.NumDC
	// maximum size of partition (excluding metadata)
	MaxPartitionSize float64 = config.MaxPartitionSize
	// file path that logs the storage
	storage_log string = config.StorageFilename
	// time interval to log storage
	logInterval time.Duration = config.LogTimeInterval
	// time interval to run service that populates hot partition
	populatingInterval time.Duration = config.PopulatingInterval
	// time interval to run service that synchronizes replica across DCs
	syncReplicaInterval time.Duration = config.SyncReplicaInterval
	// Threshold number of reads (per partition) to trigger populating
	readThreshold uint64 = config.ReadThreshold
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
	rcLock loclock.ReadCountLockMap    // Locking for ReadMap (read-count map)
	stLock loclock.StorageTableLockMap // Locking for storage table

)

// Persist storage into a log file
// TODO: add locking
func persistStorage(table *map[string]*ds.Partition) {
	// Periodically log storage
	c := time.Tick(logInterval)
	for _ = range c {
		fmt.Println("Starting to persist storage")
		// Delete old storage log
		_, e := os.OpenFile(storage_log, os.O_WRONLY|os.O_CREATE, 0755)
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
	}
}

// Write request handler
// This will automatically create a handling thread
func (l *Listener) HandleWriteReq(req ds.WriteReq, resp *ds.WriteResp) error {
	// Parse write request, get blob info
	content := req.Content
	size := req.Size
	if size > config.MaxPartitionSize {
		err := errors.New("Cannot write file of which size is bigger than partition")
		fmt.Println(err.Error())
		log.Fatal(err)
	}
	now := time.Now().Unix()
	blobUUID, err := util.NewUUID()
	if err != nil {
		fmt.Println("Error generating UUID")
		fmt.Println(err.Error())
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
			fmt.Println("Error generating UUID")
			log.Fatal(err)
		}

		// Also new entries in lock map
		rcLock.AddEntry(partitionID)
		stLock.AddEntry(partitionID)

		storageTable[partitionID] = &(ds.Partition{
			partitionID, []ds.Blob{{blobUUID, content, size, now}}, now, size})

		// Also crete new entries in replica map and read map
		ReadMap[partitionID] = &(ds.NumRead{0, 0})
		ReplicaMap[partitionID] = &(ds.PartitionState{partitionID, []string{DCID}})

	} else {
		// Add blob to partition
		stLock.Lock(partitionID)
		storageTable[partitionID].AppendBlob(ds.Blob{blobUUID, content, size, now})
		storageTable[partitionID].PartitionSize += size
		stLock.Unlock(partitionID)
	}

	// Reply with (PartitionID, blobID) pair
	*resp = ds.WriteResp{partitionID, blobUUID}

	// Print storage table after write
	if config.PrintServiceOn {
		fmt.Println("Storage Table after update:")
		util.PrintStorage(&storageTable)
		fmt.Println("------")
	}

	return nil
}

// Read request handler
// This will automatically create a handling thread
// For now, there is no not-found, because a later-finished thread might change
// resp though the first one found it
func (l *Listener) HandleReadReq(req ds.ReadReq, resp *ds.ReadResp) error {
	// Parse read request
	partitionID := req.PartitionID
	blobID := req.BlobID

	ch := make(chan bool)

	///DEBUG
	//fmt.Print("START:  ", req, "\n")

	// Look up in local storage
	go func() {
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
			ch <- true
		}
		/*
			else {
				// Not found
				*resp = ds.ReadResp{"NOT_FOUND", 0}
			}
		*/
	}()

	// Look up in other DC(s)
	for dcID, IPaddr := range IPMap {
		if dcID != DCID {
			go func(addr config.ServerIPAddr) {
				client, err := rpc.DialHTTP("tcp", addr.ServerIP+":"+addr.ServerPort1)
				if err != nil {
					fmt.Println("Dial HTTP error in handle read request. Target IP: "+addr.ServerIP, " port: ", addr.ServerPort1)
					fmt.Println(err.Error())
					log.Fatal(err)
				}

				var reply ds.ReadResp
				// Send message to other DCs, response stored in &reply
				err = client.Call("Listener.HandleRoutedReadReq", req, &reply)
				if err != nil {
					fmt.Println("RPC error in handle read request. Target IP: "+addr.ServerIP, " port: ", addr.ServerPort1)
					fmt.Println(err.Error())
					log.Fatal(err)
				}

				if reply.Size != 0 {
					*resp = reply
					ch <- true
				}
				/*
					else {
						// Not found
						*resp = ds.ReadResp{"NOT_FOUND", 0}
					}
				*/
			}(IPaddr)
		}
	}

	// DEBUG
	//fmt.Println("FINISH:  ", resp.Content, resp.Size)

	<-ch // Finish as soon as ONE result found in ANY of the DCs
	return nil
}

// Handle read request send from other DCs/servers (NOT clients)
// Separated from HandleReadRequest to avoid recursive broadcast
func (l *Listener) HandleRoutedReadReq(
	req ds.ReadReq, resp *ds.ReadResp) error {
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
// Periodically scan read-count map and make decisions independently
func populateReplica() {
	c := time.Tick(populatingInterval)
	for _ = range c {
		// Scan read map
		// TODO: sync barrier here
		for partitionID, count := range ReadMap {
			// Copy hot partitions
			// If copy-everywhere policy is turned on, always copy to all other DCs
			if count.LocalRead >= readThreshold || config.CopyEveryWhereOn {
				// Send partition to all other  DC
				for dcID, ipAddr := range IPMap {
					// Only send to DC that haven't got this partition/replica
					if dcID != DCID && !ReplicaMap[partitionID].FindDC(dcID) {
						go func(serverIP string, serverPort string, pID string, dID string) {
							fmt.Println(
								"Start populating partition : " + pID + " to DC: " + dID + " with serverIP: " + serverIP + " port: " + serverPort)
							client, err := rpc.DialHTTP("tcp", serverIP+":"+serverPort)
							if err != nil {
								fmt.Println("Dial HTTP error in populating replica. ")
								log.Fatal(err)
							}
							// Copy partition to message struct to be sent
							stLock.Lock(pID)
							var msg = *storageTable[pID]
							partitionSize := storageTable[pID].PartitionSize
							stLock.Unlock(pID)

							// Simualte transfer latency
							util.MockTransLatency(DCID, dID, partitionSize)

							var reply bool
							err = client.Call(
								"Listener.ReceivePopulatingReplica", &msg, &reply)
							if err != nil {
								fmt.Println("Error in receiving reply from populating partition:", pID)
								log.Fatal(err)
							}
							if reply {
								fmt.Println("Partition: ", pID, " populatED to DC ", dID)
							}
							// Update replica map after sending partition
							ReplicaMap[pID].AddDC(dID)
							// No need to update readmap or storage table, because already
							// have that partition (this is the sender itself)
						}(ipAddr.ServerIP, ipAddr.ServerPort1, partitionID, dcID)
					}
				}
			}
		}
	}

}

// Add partition to local storage  upon receiving another DC's populating req
// TODO: add locking here
func (l *Listener) ReceivePopulatingReplica(
	newPartition *ds.Partition, resp *bool) error {

	newPID := newPartition.PartitionID
	fmt.Println("Receiving populated partition: ", newPID)

	// TODO: get rid of this after fixing distributed sync of replica map
	if _, ok := storageTable[newPID]; ok {
		fmt.Println("Partition ", newPID, "Already exists, no merging")
		*resp = true
		return nil
	}

	// Add new entry in lockmaps
	rcLock.AddEntry(newPID)
	stLock.AddEntry(newPID)

	// Add new partition to storage table
	storageTable[newPID] = newPartition

	// Update replica map after receiving new parition
	if _, ok := ReplicaMap[newPID]; !ok {
		// New partition not in replica map, create new entry
		ReplicaMap[newPID] = &ds.PartitionState{
			newPID, []string{DCID}}
	} else {
		// New partition already in replica map, just add self-DCID to DCList
		ReplicaMap[newPID].AddDC(DCID)
	}

	// Add new entry in read map after receiving new partition
	ReadMap[newPID] = &ds.NumRead{0, 0}

	// Print storage table after adding new partition
	if config.PrintServiceOn {
		fmt.Println("Storage Table after receiving replica:")
		util.PrintStorage(&storageTable)
		fmt.Println("------")
	}

	*resp = true
	return nil
}

// Periodic service that synchronizes paritions with the same paritionID
// i.e. a replica set
// TODO: fine-grained locking
func syncReplica() {
	c := time.Tick(syncReplicaInterval)
	for _ = range c {
		// Scan replica map
		for partitionID, partitionState := range ReplicaMap {
			for _, id := range (*partitionState).DCList {
				// Sync current partition with all OTHER DC(s) that store it
				if DCID != id {
					go func(dcID string, pID string) {
						fmt.Println("Start syncing replica: " + pID + " to DC: " + dcID)
						client, err := rpc.DialHTTP(
							"tcp", IPMap[dcID].ServerIP+":"+IPMap[dcID].ServerPort1)
						if err != nil {
							fmt.Println("Dial HTTP error in sync replica.")
							fmt.Println(err.Error())
							log.Fatal(err)
						}
						// Copy partition to message (to be sent)
						stLock.Lock(pID)
						var msg = *storageTable[pID]
						stLock.Unlock(pID)

						var reply ds.Partition
						err = client.Call("Listener.ReceiveSyncReplica", &msg, &reply)
						if err != nil {
							fmt.Println("RPC error in sync partition:", pID)
							fmt.Println(err.Error())
							log.Fatal(err)
						}

						// Simualte transfer latency
						stLock.Lock(pID)
						if config.PrintServiceOn {
							fmt.Println("Simulating transfer latency")
						}
						deltaSize := util.DeltaSize(storageTable[pID], &reply)
						stLock.Unlock(pID)
						util.MockTransLatency(DCID, dcID, deltaSize)

						// Merge reply of other DCs to local partition
						stLock.Lock(pID)
						util.MergePartition(storageTable[pID], &reply)
						stLock.Unlock(pID)
					}(id, partitionID)
				}
			}
		}
	}
}

// Handler that merge a incoming partiion to local replication
// TODO: locking
func (l *Listener) ReceiveSyncReplica(
	comingPartition *ds.Partition, resp *ds.Partition) error {

	pID := comingPartition.PartitionID
	fmt.Println("Receive delta from sync partition: ", pID)

	// Store current partition to reply
	stLock.Lock(pID)
	*resp = *storageTable[pID]
	// Merge current partition and incoming partition
	util.MergePartition(storageTable[pID], comingPartition)
	stLock.Unlock(pID)

	// Print storage table after adding new partition
	if config.PrintServiceOn {
		fmt.Println("Storage Table after synchronizing replica:")
		util.PrintStorage(&storageTable)
		fmt.Println("------")
	}

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
	stLock.CreateLockMap(&storageTable)
	// Setup routing
	IPMap.CreateIPMap()
}

// Server main loop
func main() {
	// Parse DCID from command line
	if len(os.Args) != 2 {
		err := errors.New("Need one command line argument to specify DCID")
		fmt.Println(err)
		log.Fatal(err)
	}
	switch id := os.Args[1]; id {
	case "1":
		DCID = config.DC1
	case "2":
		DCID = config.DC2
	case "3":
		DCID = config.DC3
	default:
		err := errors.New(
			"Error parsing DCID from command line: need to be either 1 2 or 3")
		fmt.Println(err)
		log.Fatal(err)
	}

	fmt.Println("Storage server starts")

	// Initiates a thread that periodically persist storage on disk
	if config.LogServiceOn {
		go persistStorage(&storageTable)
	}

	// Initiates populating service
	if config.PopulateServiceOn {
		go populateReplica()
	}

	// Initiaes replica synchronization service
	if config.SyncServiceOn {
		go syncReplica()
	}

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
