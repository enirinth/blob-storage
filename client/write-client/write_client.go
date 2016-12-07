package main

/// <partition_id, blob_id, rd_access>

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io/ioutil"
	"net/rpc"
	"strconv"
	"strings"
)

func readFile() []string {
	dat, err := ioutil.ReadFile("input.txt")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	return lines
}

func main() {
	client, err := rpc.Dial("tcp", "localhost:42586")
	if err != nil {
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
		if vars[1] <= 0 {
			log.Fatal(errors.New("File size cannot be smaller or equal to zero"))
		}
		var msg = ds.WriteReq{vars[0], f}
		var reply ds.WriteResp

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
