/**********************************
* Project:  blob-storage
* Author:   Ray Chen
* Email:    raychen0411@gmail.com
* Time:     12-03-2016
* All rights reserved!
***********************************/

package main

import (
	"fmt"
	"net"
	"net/rpc"
	log "github.com/Sirupsen/logrus"
	ds "github.com/enirinth/blob-storage/clusterds"
	config "github.com/enirinth/blob-storage/clusterconfig"
	"github.com/enirinth/blob-storage/util"
	"github.com/enirinth/blob-storage/locking/loclock"
	"time"
	"net/http"
	"os"
	"errors"
)

const (
	numDC               int           = config.NumDC            // total number of DCs
	MaxPartitionSize    float64       = config.MaxPartitionSize // maximum size of partition (excluding metadata)
)

type Listener int


var (
	DCID string // ID of current DC

	// Routing
	IPMap config.ServerIPMap
	CentralIPMap config.CentralManagerIPMap

	// Cluster map data structures
	ReplicaMap = make(map[string]*ds.PartitionState)

	// Local (intra-DC) storage data structures
	storageTable = make(map[string]*ds.Partition)
	ReadMap      = make(map[string]*ds.NumRead)

	// Locking
	rcLock loclock.ReadCountLockMap // Fined-grained locking for ReadMap (read-count map)
	stLock loclock.StorageTableLockMap // Locking for storage table
)


func (l *Listener) HandleCentralManagerReadRequest(req ds.ReadReq, resp *ds.CentralManagerReadResp) error{
	partitionID := req.PartitionID
	blobID := req.BlobID

	// Look for target blob
	if _, ok := storageTable[partitionID]; ok {
		for _, blob := range storageTable[partitionID].BlobList {
			if blob.BlobID == blobID {
				dcName := util.GetRandomDCFromList(ReplicaMap[partitionID].DCList)
				address := IPMap[dcName].ServerIP + ":" + IPMap[dcName].ServerPort1
				*resp = ds.CentralManagerReadResp{Address: address, Size: blob.BlobSize}
				break
			}
		}
	}

	if resp.Size != 0 {
		rcLock.WLock(partitionID)   // Update read count
		// no need to update the global read for centralized manager
		//time.Sleep(5*time.Second)
		ReadMap[partitionID].LocalRead += 1
		rcLock.WUnlock(partitionID)
	}
	fmt.Println("read: ", req)
	return nil
}


func (l *Listener) HandleCentralManagerWriteRequest(req ds.WriteReq, resp *ds.WriteResp) error{
	// Parse write request, get blob info
	content := req.Content
	size := req.Size
	now := time.Now().Unix()
	blobUUID, err := util.NewUUID()
	if err != nil {
		log.Fatal(err)
	}

	// Select the first partition that is not full (this is different from Ambry)
	// TODO: may be slow to always iterate from the beginning
	partitionID := ""
	for id, partition := range storageTable {
		if partition.PartitionSize + size <= MaxPartitionSize {
			partitionID = id
			break
		}
	}

	var randomDC string
	if len(partitionID) == 0 {
		// If all partitions are full, create a new partition
		partitionID, err = util.NewUUID()
		if err != nil {
			log.Fatal(err)
		}

		// Also new entries in lock map
		rcLock.AddEntry(partitionID)
		stLock.AddEntry(partitionID)

		storageTable[partitionID] = &(ds.Partition{  PartitionID: partitionID,
			BlobList: []ds.Blob{{blobUUID, content, size, now}}, CreateTimestamp: now, PartitionSize : size})

		// Also crete new entries in replica map and read map
		ReadMap[partitionID] = &(ds.NumRead{ LocalRead: 0, GlobalRead: 0})

		randomDC = util.GetRandomDC(numDC)
		ReplicaMap[partitionID] = &(ds.PartitionState{PartitionID: partitionID, DCList: []string{randomDC}})

	} else {
		// Add blob to partition
		stLock.Lock(partitionID)
		//time.Sleep(5 * time.Second)
		storageTable[partitionID].AppendBlob(ds.Blob{blobUUID, content, size, now})
		storageTable[partitionID].PartitionSize += size
		stLock.Unlock(partitionID)
	}

	fmt.Println("write ", randomDC, partitionID, blobUUID)
	*resp = ds.WriteResp{PartitionID: partitionID, BlobID: blobUUID}   // Reply with (PartitionID, blobID) pair
	return nil
}


func (l *Listener) HandleCentralManagerShowStatus(req string, resp *string) error{
	util.PrintCluster(&ReplicaMap, &ReadMap)
	util.PrintStorage(&storageTable)
	return nil
}


// test lock for ReadMap
func (l * Listener) HandleCentralManagerReadLocalNum(req ds.ReadReq, resp *string) error{
	partitionID := req.PartitionID
	rcLock.RLock(partitionID)
	//time.Sleep(5 * time.Second)
	fmt.Println(ReadMap[partitionID])
	rcLock.RUnlock(partitionID)
	return nil
}


func init() {
	IPMap.CreateIPMap()
	CentralIPMap.CreateIPMap()
	rcLock.CreateLockMap(&ReadMap)
	stLock.CreateLockMap(&storageTable)
}


func testRandom() {
	for i:=0; i<20; i++ {
		random := util.GetRandomDC(3)
		fmt.Println(random)
	}
}


func main() {
	// Parse DCID from command line
	if len(os.Args) > 1 {
		err := errors.New("No para is needed! Just run the server!")
		log.Fatal(err)
	}

	address := CentralIPMap[config.DC0].ServerIP + ":" + CentralIPMap[config.DC0].ServerPort1
	fmt.Println("Central manager starts,", address)
	port := CentralIPMap[config.DC0].ServerPort1

	listener := new(Listener)
	rpc.Register(listener)
	rpc.HandleHTTP()
	inbound, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(inbound, nil)
}
