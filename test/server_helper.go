/**********************************
* Project:  blob-storage
* Author:   Ray Chen
* Email:    raychen0411@gmail.com
* Time:     12-08-2016
* All rights reserved!
***********************************/

package main

import (
	"log"
	"fmt"
	ds "github.com/enirinth/blob-storage/clusterds"
	config "github.com/enirinth/blob-storage/clusterconfig"
	"sync"
	"net/rpc"
	"os"
	"errors"
	//"github.com/enirinth/blob-storage/util"
	"strings"
	"strconv"
	"text/tabwriter"
	"io/ioutil"
)

var (
	wg sync.WaitGroup
	managerAddr string
	IPMap config.ServerIPMap
	CentralIPMap config.CentralManagerIPMap
)

func sendCentralManagerPrintReq(address string, wg *sync.WaitGroup) {
	defer wg.Done()
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	var msg string
	var reply string

	err = client.Call("Listener.HandleCentralManagerShowStatus", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Print request sent!")
}


func sendDCRequest(address string, partitionID string, blobID string, size float64, wg *sync.WaitGroup){
	defer wg.Done()

	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	var msg = ds.CentralDCReadReq{partitionID, blobID, size, "1"}
	var reply ds.CentralDCReadResp

	err = client.Call("Listener.HandleCentralDCReadRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DC:", msg, reply)
}


func sendCentralManagerReadReq(address string, partitionID string, blobID string, wg *sync.WaitGroup){
	defer wg.Done()
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}
	var msg = ds.ReadReq{partitionID, blobID}
	var reply ds.CentralManagerReadResp

	err = client.Call("Listener.HandleCentralManagerReadRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(1)
	go sendDCRequest(reply.Address, partitionID, blobID, reply.Size, wg)
}


func sendCentralManagerReadAllReq(address string, partitionID string, blobID string, wg *sync.WaitGroup){
	defer wg.Done()
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	var msg = ds.ReadReq{partitionID, blobID}
	var reply ds.CentralManagerReadResp

	err = client.Call("Listener.HandleCentralManagerReadRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(1)
	go sendDCRequest(reply.Address, partitionID, blobID, reply.Size, wg)
}


func sendCentralManagerWriteReq(address string, content string, size float64, wg *sync.WaitGroup){
	defer wg.Done()
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("Connection error", err)
	}

	var msg = ds.WriteReq{content, size}
	var reply ds.WriteResp

	err = client.Call("Listener.HandleCentralManagerWriteRequest", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PartitionID:", reply.PartitionID, " Blob:", reply.BlobID)
}


func handleShow () {
	if len(os.Args) != 2 {
		err := errors.New("Wrong input, E.g: go run show")
		log.Fatal(err)
	}
	wg.Add(1)
	go sendCentralManagerPrintReq(managerAddr, &wg)
}


func handleRead() {
	if len(os.Args) != 4 {
		err := errors.New("Wrong input, E.g: go run read partitionID blobID")
		log.Fatal(err)
	}
	partitionID := os.Args[2]
	blobID := os.Args[3]
	wg.Add(1)
	go sendCentralManagerReadReq(managerAddr, partitionID, blobID, &wg)
}

func handleMultiRead() {
	if len(os.Args) != 4 {
		err := errors.New("Wrong input, E.g: go run read partitionID blobID")
		log.Fatal(err)
	}
}


func handleReadAll() {
	if len(os.Args) != 2 {
		err := errors.New("Wrong input, E.g: go run readall")
		log.Fatal(err)
	}
	filename := "../client/read-client/central_manager_storage.txt"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	numFiles := len(lines) - 1
	for i:=0; i<numFiles; i++ {
		vars := strings.Split(lines[i], " ")
		if len(vars) != 3 {
			log.Fatal("input line error", len(vars), vars)
		}else {
			wg.Add(1)
			go sendCentralManagerReadAllReq(managerAddr, vars[0], vars[1], &wg)
		}
	}
}


func handleWrite(){
	if len(os.Args) != 4 {
		err := errors.New("Wrong input, E.g: go run server_helper.go write content size")
		log.Fatal(err)
	}
	content := os.Args[2]

	size, _ := strconv.ParseFloat(os.Args[3], 64)
	wg.Add(1)
	go sendCentralManagerWriteReq(managerAddr, content, size, &wg)
}


func handleHelp() {
	fmt.Println("#### Available commands ###")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "paras\tDescription")

	fmt.Fprintln(w, " -show \t # print server tables")
	fmt.Fprintln(w, " -read partitionID blobID \t # read file from server")
	fmt.Fprintln(w, " -readall \t # read all the files from client/read-client/central_manager_storage")
	fmt.Fprintln(w, " -write content size \t # write new file to server")

	w.Flush()
}


func handleStorageReq() {
	var hotCopySize float64
	var blindCopySize float64
	blindCopyMap := make(map[string] float64)


	for _, info := range IPMap {
		addr := info.ServerIP + ":" + info.ServerPort1
		client, err := rpc.DialHTTP("tcp", addr)

		msg := ""
		var reply ds.StorageInfo

		err = client.Call("Listener.HandleStorageInfo", msg, &reply)
		if err != nil {
			log.Fatal(err)
		}

		tmp := reply.Storage
		for k, v := range tmp {
			hotCopySize += v
			if _, ok := blindCopyMap[k]; !ok {
				blindCopyMap[k] = v
			}
		}
	}

	for _, v := range blindCopyMap {
		blindCopySize += v*3
	}

	fmt.Println("blind copy", blindCopySize)
	fmt.Println("hot copy", hotCopySize)
}


func init() {
	IPMap.CreateIPMap()
	CentralIPMap.CreateIPMap()
}


func main() {
	if len(os.Args) < 2 {
		err := errors.New("Invalid paramater input, e.g go run server_helper.go from show")
		log.Fatal(err)
	}

	managerAddr = CentralIPMap[config.DC0].ServerIP + ":" + CentralIPMap[config.DC0].ServerPort1
	fmt.Println(managerAddr)
	arg := os.Args[1]

	switch arg {
	case "show":
		handleShow()
	case "storage":
		handleStorageReq()
	case "read":
		handleRead()
	case "multiread":
		handleMultiRead()
	case "readall":
		handleReadAll()
	case "write":
		handleWrite()
	case "help":
		handleHelp()
	}
	wg.Wait()
}
