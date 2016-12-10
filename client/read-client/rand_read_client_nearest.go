package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
    config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
    "math/rand"
	"net/rpc"
    "github.com/enirinth/blob-storage/routing"
	"strconv"
	"strings"
	"sync"
	"time"
)

//const numFiles = readReqFiles
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

func readFile(input_file string) [numFiles]info {
	read_file := input_file
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

    rand.Seed(time.Now().UnixNano()) // takes the current time in nanoseconds as the seed
    input_file := os.Args[1]
	info_array := readFile(input_file)


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

    /// Testing: Printing Map
    // for key, value := range m {
    //         fmt.Println("Key:", key, "Value:", value.PartitionID, value.BlobID)
    // }

    ///Send Requests
    for i:=0; i<cnt; i++ {
        randNum := rand.Intn(cnt)
        partitionID := m[randNum].PartitionID
        blobID := m[randNum].BlobID
	    wg.Add(1)
		go sendRequest(partitionID, blobID, DCID)
        //fmt.Print(randNum, "\n")
    }
    wg.Wait()
    return
}
