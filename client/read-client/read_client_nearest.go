package main

import (
	//"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
    config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
	"net/rpc"
    //"os"
    "github.com/enirinth/blob-storage/routing"
	"strconv"
	"strings"
	"sync"
	"time"
)

const numFiles = 10

var (
    DCID string
    IPMap config.ServerIPMap
	wg sync.WaitGroup
)

type info struct {
	PartitionID string
	BlobID      string
	readReqDist int64
}


// readResponses_chan := make(chan )

func readFile() [numFiles]info {
	read_file := "out.txt"
	dat, err := ioutil.ReadFile(read_file)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")

	var info_array [numFiles]info
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

func sendRequest(PartitionID string, BlobID string, DCID string) {
	defer wg.Done()
	client, err := rpc.DialHTTP("tcp", IPMap[DCID].ServerIP+":"+IPMap[DCID].ServerPort1)
	if err != nil {
		log.Fatal(err)
	}
	// Pack message from stdin to WriteReq, initiates struct to get response
	var msg = ds.ReadReq{PartitionID, BlobID}
	var reply ds.ReadResp
	// fmt.Println(PartitionID, BlobID)

	// Send message to storage server, response stored in &reply
	t0 := time.Now()
	err = client.Call("Listener.HandleReadReq", msg, &reply)
	t1 := time.Now()

	if err != nil {
        fmt.Print("ERROR: ", err.Error(), "\n")
		log.Fatal(err)
	}
	fmt.Println(msg, reply, t1.Sub(t0))

	return
}

func init() {
    IPMap.CreateIPMap()
}
func main() {
    // Select nearest DC to send request
    DCID = routing.NearestDC()

	info_array := readFile()
	fmt.Print(numFiles, "\n")
	for i := 0; i < numFiles; i++ {
		num_req_left := info_array[i].readReqDist
		for num_req_left > 0 {
			wg.Add(1)
			go sendRequest(info_array[i].PartitionID, info_array[i].BlobID, DCID)
			num_req_left -= 1
		}
	}
	wg.Wait()
	//fmt.Println(test_cnt)
}
