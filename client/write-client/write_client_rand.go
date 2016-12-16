package main

import (
	"errors"
	"fmt"
	"github.com/enirinth/blob-storage/util"
	log "github.com/Sirupsen/logrus"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
    "math/rand"
	"net/rpc"
	"os"
	"strconv"
	"strings"
    "time"
)

const FILEPATH = "../../data/"

var (
	IPMap config.ServerIPMap
	CentralIPMap config.CentralManagerIPMap
)


//func sendWriteRequest(Content string, Size float64, DCID string) (string, string){
//    address := IPMap[DCID].ServerIP + ":" + IPMap[DCID].ServerPort1
//
//    client, err := rpc.DialHTTP("tcp", address)
//	if err != nil {
//		log.Fatal(err)
//	}
//    var msg = ds.WriteReq{Content, Size}
//    var reply ds.WriteResp
//
//    err = client.Call("Listener.HandleWriteReq", msg, &reply)
//    if err != nil {
//        log.Fatal(err)
//    }
//    return reply.PartitionID, reply.BlobID
//}

func writeBlob() {
	filename := os.Args[1]
    outputFile := "out.txt"

	dat, err := ioutil.ReadFile(FILEPATH + filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	numFiles := len(lines) - 1

	addr1 := IPMap[config.DC1].ServerIP + ":" + IPMap[config.DC1].ServerPort1
	client1, err := rpc.DialHTTP("tcp", addr1)
	addr2 := IPMap[config.DC2].ServerIP + ":" + IPMap[config.DC2].ServerPort1
	client2, err := rpc.DialHTTP("tcp", addr2)
	addr3 := IPMap[config.DC3].ServerIP + ":" + IPMap[config.DC3].ServerPort1
	client3, err := rpc.DialHTTP("tcp", addr3)
	if err != nil {
		log.Fatal(err)
	}

	writeStr := ""
	for i := 0; i < numFiles; i++ {
		vars := strings.Split(lines[i], " ")
		f, err := strconv.ParseFloat(vars[1], 64)
        if err != nil {
            log.Fatal(err)
        }
		if f <= 0 {
			log.Fatal(errors.New("File size cannot be smaller or equal to zero"))
		}

        /// select random DC
        randDCID := util.GetRandomDC(3)
        fmt.Println(randDCID)

		var msg = ds.WriteReq{vars[0], f}
		var reply ds.WriteResp

		if randDCID == config.DC1 {
			err = client1.Call("Listener.HandleWriteReq", msg, &reply)
		} else if randDCID == config.DC2 {
			err = client2.Call("Listener.HandleWriteReq", msg, &reply)
		} else if randDCID == config.DC3 {
			err = client3.Call("Listener.HandleWriteReq", msg, &reply)
		}

		fmt.Println(reply.PartitionID + " " + reply.BlobID + " " + vars[2])
		curLineStr := reply.PartitionID + " " + reply.BlobID + " " + vars[2] + "\n"
		writeStr += curLineStr
	}

	d1 := []byte(writeStr)
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
    rand.Seed(time.Now().UnixNano()) // takes the current time in nanoseconds as the seed

	if len(os.Args) != 2 {
		err := errors.New("Wrong input, E.g: go run central_manager_write_blob.go input100.txt")
		log.Fatal(err)
	}
	writeBlob()
}
