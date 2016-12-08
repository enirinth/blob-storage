/**********************************
* Project:  blob-storage
* Author:   Ray Chen
* Email:    raychen0411@gmail.com
* Time:     12-07-2016
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
	"net/http"
	"os"
	"errors"
	//"github.com/enirinth/blob-storage/locking/loclock"
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
	//rcLock loclock.ReadCountLockMap // Fined-grained locking for ReadMap (read-count map)
)

type Listener int


func (l *Listener) HandleCentralDCReadRequest(req ds.CentralDCReadReq, resp *ds.CentralDCReadResp) error{
	//size := req.Size
	// TODO:  sleep for some time based on the size
	fmt.Println("Handle req", req)
	*resp = ds.CentralDCReadResp{Content: "blob content"}
	return nil
}


func init() {
	IPMap.CreateIPMap()
}


func main() {
	// Parse DCID from command line
	if len(os.Args) != 2 {
		err := errors.New("Need one command line argument to specify DCID")
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
		err := errors.New( "Error parsing DCID from command line: need to be either 1 2 or 3")
		log.Fatal(err)
	}
	port := IPMap[DCID].ServerPort1
	fmt.Println("Storage server starts:", DCID, ", ", IPMap[DCID].ServerIP +":" + port)

	listener := new(Listener)
	rpc.Register(listener)
	rpc.HandleHTTP()
	inbound, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(inbound, nil)
}