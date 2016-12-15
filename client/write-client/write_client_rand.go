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


func sendWriteRequest(Content string, Size float64, DCID string) (string, string){
    address := IPMap[DCID].ServerIP + ":" + IPMap[DCID].ServerPort1

    client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
    var msg = ds.WriteReq{Content, Size}
    var reply ds.WriteResp

    err = client.Call("Listener.HandleWriteReq", msg, &reply)
    if err != nil {
        log.Fatal(err)
    }
    return reply.PartitionID, reply.BlobID
}

func writeBlob() {
	filename := os.Args[1]
    outputFile := "out.txt"

	dat, err := ioutil.ReadFile(FILEPATH + filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	numFiles := len(lines) - 1

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

		var reply ds.WriteResp
        reply.PartitionID, reply.BlobID = sendWriteRequest(vars[0], f, randDCID)

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
