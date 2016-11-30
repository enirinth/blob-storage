package main

import (
	"bufio"
	"fmt"
	ds "github.com/enirinth/read-clock/clusterds"
	"log"
	"net/rpc"
	"os"
	"strings"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:42586")
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
