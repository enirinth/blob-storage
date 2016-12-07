package main

/// <partition_id, blob_id, rd_access>

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
)
var (
	DCID string
	IPMap config.ServerIPMap
)

func readFile() []string {
	dat, err := ioutil.ReadFile("input.txt")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	return lines
}

func init() {
    IPMap.CreateIPMap()
}

func main() {

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
	//client, err := rpc.Dial("tcp", "localhost:42011")
	if err != nil {
        //fmt.Print("HERE\n")
		log.Fatal(err)
	}
	/// Declare Write ID File

	/// Read Write Text File
	lines := readFile()

	numFiles := len(lines) - 1
	write_str := ""

	for i := 0; i < numFiles; i++ {
		vars := strings.Split(lines[i], " ")
		f, err := strconv.ParseFloat(vars[1], 64)
		if f <= 0 {
			log.Fatal(errors.New("File size cannot be smaller or equal to zero"))
		}

        var msg = ds.WriteReq{vars[0], f}
		var reply ds.WriteResp

        //fmt.Print(msg.partitionID, msg.blobID)

		err = client.Call("Listener.HandleWriteReq", msg, &reply)
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(reply)
		fmt.Println(reply.PartitionID + " " + reply.BlobID + " " + vars[2])

		cur_line_str := reply.PartitionID + " " + reply.BlobID + " " + vars[2] + "\n"
		write_str += cur_line_str
	}
	d1 := []byte(write_str)
	err = ioutil.WriteFile("../read-client/out.txt", d1, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
