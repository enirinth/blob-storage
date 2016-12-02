package main

import (
	"bufio"
	"errors"
	"fmt"
	ds "github.com/enirinth/blob-storage/clusterds"
	"log"
	"net/rpc"
	"os"
	"strconv"
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
