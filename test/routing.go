/**********************************
* Project:  blob-storage
* Author:   Ray Chen
* Email:    raychen0411@gmail.com
* Time:     12-08-2016
* All rights reserved!
***********************************/

package main

import (
	"os"
	"errors"
	"log"
	"fmt"
	"text/tabwriter"
	"github.com/enirinth/blob-storage/routing"
	"strconv"
)


func handleRoutingHelp() {
	fmt.Println("#### Available commands ###")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "paras\tDescription")

	fmt.Fprintln(w, " -latency :number \t # increase server latency")
	fmt.Fprintln(w, " -bandwidth :dc_name reduced bandwidth in kbit \t # reduce server bandwidth")

	w.Flush()
}


func main() {
	if len(os.Args) < 2 {
		err := errors.New("Invalid paramater input, run go routing.go help")
		log.Fatal(err)
	}

	arg := os.Args[1]
	switch arg {
	case "latency":
		if len(os.Args) != 3 {
			err := errors.New("Invalid input, latency is based on ms, eg: go run routing.go latency 50")
			log.Fatal(err)
		}
		latency, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			log.Fatal(err)
		}
		routing.ChangeLatency(latency)
	case "bandwidth":
		if len(os.Args) != 4 {
			err := errors.New("Input bandwidth decrease percentage, eg: go run routing.go bandwidth 1 50")
			log.Fatal(err)
		}
		dc := os.Args[2]
		bandwidth, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println(err.Error())
			log.Fatal(err)
		}
		routing.ChangeBandwidth(dc, bandwidth)
	case "help":
		handleRoutingHelp()
	default:
		fmt.Println("Wrong input, run go routing.go help")
	}
}


