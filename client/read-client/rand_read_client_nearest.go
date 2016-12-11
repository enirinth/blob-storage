package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
    config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
	"math/rand"
	"net/rpc"
    "os"
    "github.com/enirinth/blob-storage/routing"
	"strconv"
	"strings"
	"sync"
	"time"
)

const numFiles = 200

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
	//read_file := "out.txt"

	read_file := os.Args[1]
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

func sendRequest(PartitionID string, BlobID string, DCID string, client *rpc.Client ) {
	defer wg.Done()
	// Pack message from stdin to WriteReq, initiates struct to get response
	var msg = ds.ReadReq{PartitionID, BlobID}
	var reply ds.ReadResp
	// fmt.Println(PartitionID, BlobID)

	// Send message to storage server, response stored in &reply
	t0 := time.Now()
    err := client.Call("Listener.HandleReadReq", msg, &reply)
	t1 := time.Now()

	if err != nil {
        fmt.Println("Error ListenerHandlerReadReq")
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
	if len(os.Args) !=2 {
		err := errors.New("Wrong input, E.g: go run rand_read_client_nearest.go out.txt")
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
	info_array := readFile()

    // Select nearest DC to send request
    DCID = routing.NearestDC()

	client, err := rpc.DialHTTP("tcp", IPMap[DCID].ServerIP+":"+IPMap[DCID].ServerPort1)
	if err != nil {
        fmt.Println("Error rpc.DialHTTP")
		log.Fatal(err)
	}


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
    for i:=0; i<cnt; i++ {
        randNum := rand.Intn(cnt)
        partitionID := m[randNum].PartitionID
        blobID := m[randNum].BlobID
        wg.Add(1)
        go sendRequest(partitionID, blobID, DCID, client)
        //fmt.Print(randNum, "\n")
    }

    wg.Wait()
	return
}
