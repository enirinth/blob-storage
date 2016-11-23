package main

import (
	"fmt"
	ds "github.com/enirinth/read-clock/clusterds"
	util "github.com/enirinth/read-clock/util"
	"log"
	"net"
	"net/rpc"
	"time"
)

const NumDC int = 3

type Listener int

// Cluster map data structures
var ReplicaMap = make(map[string]ds.PartitionState)

var ReadMap = make(map[string]ds.NumRead)

// Local (intra-DC) storage data structures
var StorageTable = make(map[string]*ds.Partition)

func PrintStorage() {
	for _, v := range StorageTable {
		fmt.Println(*v)
	}
}

// Client request handler
func (l *Listener) GetLine(msg ds.WriteMsg, ack *bool) error {
	content := msg.Content
	size := msg.Size
	now := time.Now().Unix()
	uuid, err := util.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Received server request blob with content: " + content)

	StorageTable["1"].AppendBlob(ds.Blob{uuid, content, size, now})

	fmt.Println("Storage Table after update:")
	PrintStorage()
	fmt.Println("------")
	return nil
}

// Server main loop
func main() {
	fmt.Println("Storage server starts")
	fmt.Println("Initialize storage with dummy data")
	ReplicaMap["1"] = ds.PartitionState{"1", []string{"DC_A"}}
	ReadMap["1"] = ds.NumRead{0, 0}
	now := time.Now().Unix()
	uuid := util.NewUUID()
	StorageTable["1"] = &(ds.Partition{"1", []ds.Blob{{uuid, "Donald Trump is elected the President", 1.2, now}}, now, 1.2})
	fmt.Println("Initial ReadMap:")
	fmt.Println(ReadMap)
	fmt.Println("-------")
	fmt.Println("Initial StorageTable:")
	PrintStorage()
	fmt.Println("-------")
	// Main loop
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
	if err != nil {
		log.Fatal(err)
	}
	inbound, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)
}
