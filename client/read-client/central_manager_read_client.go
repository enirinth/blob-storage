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
	"log"
	"net/rpc"
	ds "github.com/enirinth/blob-storage/clusterds"
	"sync"
	"time"
	config "github.com/enirinth/blob-storage/clusterconfig"
	"github.com/enirinth/blob-storage/util"
	"strings"
)

var (
	DCID string
	IPMap config.ServerIPMap
)


func sendDCRequest(address string, partitionID string, blobID string, size float64, wg *sync.WaitGroup){
	defer wg.Done()

	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	// Pack message from stdin to WriteReq, initiates struct to get response
	var msg = ds.CentralDCReadReq{partitionID, blobID, size}
	var reply ds.CentralDCReadResp

	err = client.Call("Listener.HandleCentralDCReadRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DC:", msg, reply)
}


func sendCentralManagerRequest(address string, partitionID string, blobID string, wg *sync.WaitGroup) {
	defer wg.Done()
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	// Pack message from stdin to WriteReq, initiates struct to get response
	var msg = ds.ReadReq{partitionID, blobID}
	var reply ds.CentralManagerReadResp

	err = client.Call("Listener.HandleCentralManagerReadRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Manager:", msg, reply)

	wg.Add(1)
	go sendDCRequest(reply.Address, partitionID, blobID, reply.Size, wg)
}


func init() {
	IPMap.CreateIPMap()
}


func main() {
	fmt.Println("start client");
	managerAddr := IPMap[config.DC0].ServerIP + ":" + IPMap[config.DC0].ServerPort1

	//TODO: generate read requests that follow the zipf distribution
	filename := "central_manager_storage.txt"
	lines := util.ReadFile(filename)
	numFiles := len(lines) - 1
	fmt.Println(numFiles, DCID, managerAddr)

	t0 := time.Now()
	var wg sync.WaitGroup
	for i:=0; i<numFiles; i++ {
		vars := strings.Split(lines[i], " ")
		if len(vars) != 3 {
			log.Fatal("input line error", len(vars), vars)
		}else {
			wg.Add(1)
			go sendCentralManagerRequest(managerAddr, vars[0], vars[1], &wg)
		}
	}
	wg.Wait()
	t1 := time.Now()
	fmt.Println("Total time: ", t1.Sub(t0))
}
