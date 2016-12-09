package routing

import (
	"fmt"
	config "github.com/enirinth/blob-storage/clusterconfig"
	"github.com/enirinth/blob-storage/util"
	"github.com/tatsushid/go-fastping"
	"math"
	"net"
	"strconv"
	"time"
	"strings"
	"os/exec"
	"bytes"
)

// Ping ip adress (ICMP), get response time
func RespTime(ipAddr string) float64 {
	pinger := fastping.NewPinger()

	_, err := pinger.Network("udp")
	if err != nil {
		panic("Error setting network type: " + err.Error())
	}

	var t float64

	addr, err := net.ResolveIPAddr("ip", ipAddr)
	if err != nil {
		panic("Error resolving IP Address: " + err.Error())
	}

	pinger.AddIPAddr(addr)
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		t = rtt.Seconds()
	}

	if err = pinger.Run(); err != nil {
		panic(err)
	}
	return t
}

// Find nearest DC, return DCID (either 1, 2 or 3)
func NearestDC() string {
	var dcID string
	fmt.Println("Initializing...determing which DC is the nearest...")
	t1 := RespTime(config.SERVER1_IP)
	t2 := RespTime(config.SERVER2_IP)
	t3 := RespTime(config.SERVER3_IP)
	fmt.Println("Response time pinging DC1 : " + strconv.FormatFloat(t1, 'f', -1, 64))
	fmt.Println("Response time pinging DC2 : " + strconv.FormatFloat(t2, 'f', -1, 64))
	fmt.Println("Response time pinging DC3 : " + strconv.FormatFloat(t3, 'f', -1, 64))
	min := math.Min(math.Min(t1, t2), t3)
	if util.FloatEquals(t1, min) {
		dcID = "1"
	} else if util.FloatEquals(t2, min) {
		dcID = "2"
	} else if util.FloatEquals(t3, min) {
		dcID = "3"
	}
	fmt.Println("DC " + dcID + " is the nearest DC, to which all requests will be sent")
	return dcID
}


// generate shell command string for traffic control
func getTCCmdStr(latency int)  string{
	cmd := "sudo tc qdisc add dev eth0 root netem delay " + strconv.Itoa(latency) + "ms"
	return cmd
}

// execute shell command
func execCmd(cmdStr string) string {
	parts := strings.Fields(cmdStr)
	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return " "
	}
	fmt.Println("Result: " + out.String())
	return out.String()
}

// remove current traffic control setting
func clearTC() {
	cmd := "sudo tc qdisc delete dev lo root netem"
	execCmd(cmd)
}

// put latency on linux network card, eth0
func ChangeTC(latency int) {
	cmd := getTCCmdStr(latency)
	fmt.Println("update: ", cmd)
	clearTC()
	execCmd(cmd)
}
