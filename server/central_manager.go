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
	// Cluster map data structures
	ReplicaMap = make(map[string]*ds.PartitionState)

	// Local (intra-DC) storage data structures
	storageTable = make(map[string]*ds.Partition)
	ReadMap      = make(map[string]*ds.NumRead)

	// Locking
	rcLock loclock.ReadCountLockMap // Fined-grained locking for ReadMap (read-count map)
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

	// TODO: add lock when update read counter
	if resp.Size != 0 {
		//rcLock.WLock(partitionID)   // Update read count
		//ReadMap[partitionID].GlobalRead += 1
		ReadMap[partitionID].LocalRead += 1
		//rcLock.WUnlock(partitionID)
		//wg.Done()
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

	// TODO: add lock when update those global data structures
	var randomDC string
	if len(partitionID) == 0 {
		// If all partitions are full, create a new partition
		partitionID, err = util.NewUUID()
		if err != nil {
			log.Fatal(err)
		}
		storageTable[partitionID] = &(ds.Partition{
			partitionID, []ds.Blob{{blobUUID, content, size, now}}, now, size})

		// Also crete new entries in replica map and read map
		ReadMap[partitionID] = &(ds.NumRead{0, 0})

		randomDC = util.GetRandomDC(numDC)
		ReplicaMap[partitionID] = &(ds.PartitionState{partitionID, []string{randomDC}})

		// Also create new entries in lock map
		//rcLock.AddEntry(partitionID)
		//stLock.AddEntry(partitionID)
	} else {
		// Add blob to partition
		//stLock.Lock(partitionID)
		storageTable[partitionID].AppendBlob(ds.Blob{blobUUID, content, size, now})
		storageTable[partitionID].PartitionSize += size
		//stLock.Unlock(partitionID)
	}

	fmt.Println("write finish", randomDC, partitionID, blobUUID)
	*resp = ds.WriteResp{PartitionID: partitionID, BlobID: blobUUID}   // Reply with (PartitionID, blobID) pair
	return nil
}


func (l *Listener) HandleCentralManagerShowStatus(req string, resp *string) error{
	util.PrintCluster(&ReplicaMap, &ReadMap)
	util.PrintStorage(&storageTable)
	return nil
}


func init() {
	IPMap.CreateIPMap()
}


func testRandom() {
	for i:=0; i<20; i++ {
		random := util.GetRandomDC(3)
		fmt.Println(random)
	}
}


func main() {
	// Parse DCID from command line
	if len(os.Args) != 2 {
		err := errors.New("Need one command line argument to specify DCID")
		log.Fatal(err)
	}
	switch id := os.Args[1]; id {
	case "0":
		DCID = config.DC0
	case "1":
		DCID = config.DC1
	case "2":
		DCID = config.DC2
	case "3":
		DCID = config.DC3
	default:
		err := errors.New( "Error parsing DCID from command line: need to be either 1 2 or 3")
		log.Fatal(err)
	}
	fmt.Println("Storage server starts")

	port := IPMap[DCID].ServerPort1

	listener := new(Listener)
	rpc.Register(listener)
	rpc.HandleHTTP()
	inbound, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(inbound, nil)
}
