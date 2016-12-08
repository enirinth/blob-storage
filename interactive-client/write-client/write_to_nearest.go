package main

import (
	"bufio"
	"errors"
	"fmt"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"github.com/enirinth/blob-storage/routing"
	"log"
	"math"
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

var EPSILON float64 = 0.00000001

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}

func main() {
	// Find the nearest (smallest latency)  DC/server to write to
	fmt.Println("Initializing...determing which DC is the nearest...")

	t1 := routing.RespTime(config.SERVER1_IP)
	t2 := routing.RespTime(config.SERVER2_IP)
	t3 := routing.RespTime(config.SERVER3_IP)
	fmt.Println("Response time to DC1 : " + strconv.FormatFloat(t1, 'f', -1, 64))
	fmt.Println("Response time to DC1 : " + strconv.FormatFloat(t2, 'f', -1, 64))
	fmt.Println("Response time to DC1 : " + strconv.FormatFloat(t3, 'f', -1, 64))
	min := math.Min(math.Min(t1, t2), t3)
	if floatEquals(t1, min) {
		DCID = "1"
	} else if floatEquals(t2, min) {
		DCID = "2"
	} else if floatEquals(t3, min) {
		DCID = "3"
	}
	fmt.Println("DC " + DCID + " is the nearest DC, to which all writes will be sent")

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
