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
	"strconv"
	"strings"
	"sync"
	"time"
)

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

const FILEPATH = "../../data/"
// readResponses_chan := make(chan )

func readFile() []info {
	read_file := "out.txt"
	dat, err := ioutil.ReadFile(FILEPATH + read_file)
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


	//DCID = util.GetRandomDC(config.NumDC+1)    // NumDC=3, thus random from [1,4)
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
	fmt.Println(t1.Sub(t0), reply.Size)

	return
}

func init() {
    IPMap.CreateIPMap()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println()
		err := errors.New("Wrong input, E.g: go run rand_read_client.go 1 10")
		log.Fatal(err)
	}

    rand.Seed(time.Now().UnixNano()) // takes the current time in nanoseconds as the seed
	info_array := readFile()
	readNum, _ := strconv.Atoi(os.Args[2])

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
    for i:=0; i<readNum; i++ {
        randNum := rand.Intn(cnt)
        partitionID := m[randNum].PartitionID
        blobID := m[randNum].BlobID
	    wg.Add(1)
		go sendRequest(partitionID, blobID)
        //fmt.Print(randNum, "\n")
    }
    wg.Wait()
    return
}
