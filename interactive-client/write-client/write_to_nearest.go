package main

import (
	"bufio"
	"errors"
	"fmt"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"github.com/enirinth/blob-storage/routing"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

var (
	DCID string // target DCID
	// Routing
	IPMap config.ServerIPMap
)

func init() {
	// Setup routing
	IPMap.CreateIPMap()
}

func main() {
	// Find the nearest (smallest latency)  DC/server to write to
	DCID = routing.NearestDC()
	client, err := rpc.DialHTTP(
		"tcp", IPMap[DCID].ServerIP+":"+IPMap[DCID].ServerPort1)
	if err != nil {
		log.Fatal(err)
	}

	in := bufio.NewReader(os.Stdin)
	for {
		// Parse stdin
		line, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		words := strings.Fields(line)
		size, err := strconv.ParseFloat(words[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		if size <= 0 {
			log.Fatal(errors.New("File size cannot be smaller or equal to zero"))
		}
		content := words[0]

		// Pack message from stdin to WriteReq, initiates struct to get response
		var msg = ds.WriteReq{content, size}
		var reply ds.WriteResp

		// Send message to storage server, response stored in &reply
		err = client.Call("Listener.HandleWriteReq", msg, &reply)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(reply)
	}
}
