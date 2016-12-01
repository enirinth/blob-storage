package main

import (
	// "bufio"
	"fmt"
	ds "github.com/enirinth/blob-storage/clusterds"
	"log"
	"net/rpc"
	// "os"
    "io/ioutil"
    "strconv"
	"strings"
)


type info struct {
    PartitionID string
    BlobID      string
    readReqDist int64
}

func readFile() [5]info {
    read_file := "out.txt"
    dat, err := ioutil.ReadFile(read_file)
    if err != nil {
        log.Fatal(err)
    }
    lines := strings.Split(string(dat), "\n")

    const size = 5

    var info_array [size] info
    for i := 0; i < size; i++ {
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
    client, err := rpc.Dial("tcp", "localhost:42586")
    if err != nil {
    	log.Fatal(err)
    }
    // Pack message from stdin to WriteReq, initiates struct to get response
    var msg = ds.ReadReq{PartitionID, BlobID}
	var reply ds.ReadResp

    fmt.Println(PartitionID, BlobID)

	// Send message to storage server, response stored in &reply
	err = client.Call("Listener.HandleReadReq", msg, &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)

    return
}

func main() {
	// client, err := rpc.Dial("tcp", "localhost:42586")
	// if err != nil {
	// 	log.Fatal(err)
	// }

    /// Read p_id, b_id, rd_req
    info_array := readFile()
    for i:=0; i<5; i++ {
        // fmt.Println(info_array[i].PartitionID, info_array[i].BlobID)
        sendRequest(info_array[i].PartitionID, info_array[i].BlobID)
    }




}
