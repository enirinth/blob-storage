package main

import (
	"bufio"
	"fmt"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"log"
	"net/rpc"
	"os"
	"strings"
)

func main() {
	// TODO: select nearest DC to send request
	client, err := rpc.DialHTTP("tcp", config.SERVER1_IP+":"+config.SERVER1_PORT1)
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
		partitionID := words[0]
		blobID := words[1]

		// Pack message from stdin to WriteReq, initiates struct to get response
		var msg = ds.ReadReq{partitionID, blobID}
		var reply ds.ReadResp

		// Send message to storage server, response stored in &reply
		err = client.Call("Listener.HandleReadReq", msg, &reply)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(reply)
	}

}
