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
	"strings"
	"math/rand"
	"io/ioutil"
	"strconv"
	"os"
	"errors"
)

const FILEPATH = "../../data/"

var (
	DCID string
	IPMap config.ServerIPMap
	CentralIPMap config.CentralManagerIPMap
	wg sync.WaitGroup
	t0 time.Time
)

type info struct {
	PartitionID string
	BlobID      string
	readReqDist int64
}


func readCentralFile(filename string) []info {
	dat, err := ioutil.ReadFile(FILEPATH + filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	numFiles := len(lines) - 1

	info_array := make([]info, numFiles)
	for i := 0; i < numFiles; i++ {
		x := strings.Split(lines[i], " ")
		info_array[i].PartitionID = x[0]
		info_array[i].BlobID = x[1]

		readReqDist, err := strconv.ParseInt(x[2], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		info_array[i].readReqDist = readReqDist
	}
	return info_array
}

func sendDCRequest(address string, partitionID string, blobID string, size float64, wg *sync.WaitGroup){
	defer wg.Done()

	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	// Pack message from stdin to WriteReq, initiates struct to get response
	var msg = ds.CentralDCReadReq{partitionID, blobID, size, DCID}
	var reply ds.CentralDCReadResp

	err = client.Call("Listener.HandleCentralDCReadRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}
	t1 := time.Now()
	fmt.Println(t1.Sub(t0))
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

	wg.Add(1)
	go sendDCRequest(reply.Address, partitionID, blobID, reply.Size, wg)
}


func init() {
	CentralIPMap.CreateIPMap()
	IPMap.CreateIPMap()
}


func main() {
	if len(os.Args) != 3 {
		fmt.Println()
		err := errors.New("Wrong input, E.g: go run central_manager_read_client.go 1 10")
		log.Fatal(err)
	}
	//fmt.Println("start client");
	DCID = os.Args[1]
	managerAddr := CentralIPMap[config.DC0].ServerIP + ":" + CentralIPMap[config.DC0].ServerPort1
	readNum, _ := strconv.Atoi(os.Args[2])

	t0 = time.Now()
	filename := "central_manager_storage.txt"
	info_array := readCentralFile(filename)

	/// Create map of num => {part_id, blob_id}
	m := make(map[int]info)
	cnt := 0
	for i:=0; i<len(info_array); i++ {
		num_req_left := int(info_array[i].readReqDist)
		for j:=0; j<num_req_left; j++ {
			m[cnt] = info_array[i]
			cnt += 1
		}
	}

	///Send Requests
	for i:=0; i<readNum; i++ {
		randNum := rand.Intn(cnt)
		partitionID := m[randNum].PartitionID
		blobID := m[randNum].BlobID
		wg.Add(1)
		go sendCentralManagerRequest(managerAddr, partitionID, blobID, &wg)
	}
	wg.Wait()
}
