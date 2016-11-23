package main

import (
	"bufio"
	ds "github.com/enirinth/read-clock/clusterds"
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
		// Parse message from stdin
		line, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		words := strings.Fields(line)
		size, err := strconv.ParseFloat(words[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		content := words[0]
		var msg = ds.WriteMsg{content, size}
		var reply bool
		// Send message
		err = client.Call("Listener.GetLine", msg, &reply)
		if err != nil {
			log.Fatal(err)
		}
	}
}
