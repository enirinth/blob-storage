package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"fmt"
)

const FILEPATH = "../../data/"

var (
	IPMap config.ServerIPMap
	CentralIPMap config.CentralManagerIPMap
)


func writeBlob(address string) {
	filename := os.Args[1]
	serverCall := ""
	outputFile := ""

	serverCall = "Listener.HandleCentralManagerWriteRequest"
	outputFile = "central_manager_storage.txt"

	dat, err := ioutil.ReadFile(FILEPATH + filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	numFiles := len(lines) - 1

	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	writeStr := ""
	for i := 0; i < numFiles; i++ {
		vars := strings.Split(lines[i], " ")
		f, err := strconv.ParseFloat(vars[1], 64)
		if f <= 0 {
			log.Fatal(errors.New("File size cannot be smaller or equal to zero"))
		}

		var msg = ds.WriteReq{vars[0], f}
		var reply ds.WriteResp

		err = client.Call(serverCall, msg, &reply)
		if err != nil {
			log.Fatal(err)
		}

		curLineStr := reply.PartitionID + " " + reply.BlobID + " " + vars[2] + "\n"
		writeStr += curLineStr
	}
	d1 := []byte(writeStr)
	fmt.Println(FILEPATH+outputFile)
	err = ioutil.WriteFile(FILEPATH + outputFile, d1, 0644)
	if err != nil {
		log.Fatal(err)
	}
}


func init() {
	IPMap.CreateIPMap()
	CentralIPMap.CreateIPMap()
}


func main() {
	if len(os.Args) != 2 {
		err := errors.New("Wrong input, E.g: go run central_manager_write_blob.go input.txt")
		log.Fatal(err)
	}
	var address string
	dc := config.DC0
	address = CentralIPMap[dc].ServerIP + ":" + CentralIPMap[dc].ServerPort1
	fmt.Println(address)
	writeBlob(address)
}
