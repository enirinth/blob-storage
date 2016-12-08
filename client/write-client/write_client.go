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
	"github.com/enirinth/blob-storage/util"
)

var (
	DCID  string
	IPMap config.ServerIPMap
)


func writeBlob(address string) {
	dcid := os.Args[1]
	filename := os.Args[2]
	serverCall := ""
	outputFile := ""

	if dcid == config.DC0 {
		serverCall = "Listener.HandleCentralManagerWriteRequest"
		outputFile = "central_manager_storage.txt"
	}else {
		serverCall = "Listener.HandleWriteReq"
		outputFile = "out.txt"
	}
	lines := util.ReadFile(filename)
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

		fmt.Println(reply.PartitionID + " " + reply.BlobID + " " + vars[2])
		curLineStr := reply.PartitionID + " " + reply.BlobID + " " + vars[2] + "\n"
		writeStr += curLineStr
	}
	d1 := []byte(writeStr)
	err = ioutil.WriteFile("../read-client/" + outputFile, d1, 0644)
	if err != nil {
		log.Fatal(err)
	}
}


func init() {
    IPMap.CreateIPMap()
}


func main() {
	if len(os.Args) != 3 {
		err := errors.New("Wrong input, E.g: go run write_client.go 0 input100.txt")
		log.Fatal(err)
	}
	dc := os.Args[1]
	address := IPMap[dc].ServerIP + ":" + IPMap[dc].ServerPort1
	fmt.Println(address)
	writeBlob(address)
}
