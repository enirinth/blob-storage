package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
    config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
	"net/rpc"
    "os"
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

func sendRequest(PartitionID string, BlobID string) {
	defer wg.Done()

    // Parse DCID from command line
    switch id := os.Args[1]; id {
    case "1":
        DCID = config.DC1
    case "2":
        DCID = config.DC2
    case "3":
        DCID = config.DC3
    default:
        log.Fatal(errors.New("Error parsing DCID from command line"))
    }


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
		fmt.Print("ERROR\n")
		log.Fatal(err)
	}
	fmt.Println(msg, reply, t1.Sub(t0))

	return
}

func init() {
    IPMap.CreateIPMap()
}
func main() {
	info_array := readFile()
//	test_cnt := 0
	fmt.Print(numFiles, "\n")
	for i := 0; i < numFiles; i++ {
		num_req_left := info_array[i].readReqDist
		for num_req_left > 0 {
			// fmt.Println(info_array[i].PartitionID + " " + info_array[i].BlobID + "\n")
			wg.Add(1)
			go sendRequest(info_array[i].PartitionID, info_array[i].BlobID)
			num_req_left -= 1
			//test_cnt += 1
		}
	}
	fmt.Print("HERE0")
	wg.Wait()
	fmt.Print("HERE1")
	//fmt.Println(test_cnt)
}
