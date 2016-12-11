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
	fmt.Println("### Result ###")
	fmt.Print(out.String())
	return out.String()
}

// remove current traffic control setting on eth0
func ClearTC() {
	cmd := "sudo tc qdisc delete dev eth0 root"
	fmt.Println("Clear TC setting: ", cmd)
	execCmd(cmd)
}

// show current traffic control setting
func ShowTC() {
	cmd := "tc qdisc show"
	execCmd(cmd)
}

// put latency on linux network card, eth0
func ChangeLatency(latency int) {
	ClearTC()

	cmd := "sudo tc qdisc add dev eth0 root netem delay " + strconv.Itoa(latency) + "ms"
	fmt.Println("Update TC setting: ", cmd)
	execCmd(cmd)
}

// use tbf(token bucket filter) to change the throughput, setting bandwidth*=(100-percentage)/100
// burst output is set by default, drop packages when wait time longer than latency.
// bandwidth is based on kbit, not kbyte!!
// input percentage is [0, 100]
// default burst is 50000kb = 50MB
// default latency is 500ms
func ChangeBandwidth(dc string, percentage int) {
	ClearTC()
	var defaultBandWidth int
	var newBandwidth string
	switch dc {
	case config.DC0:
		defaultBandWidth = config.BANDWIDTH0 * 1000
	case config.DC1:
		defaultBandWidth = config.BANDWIDTH1 * 1000
	case config.DC2:
		defaultBandWidth = config.BANDWIDTH2 * 1000
	case config.DC3:
		defaultBandWidth = config.BANDWIDTH3 * 1000
	}

	fmt.Println("Default bandwidth on dc = " + dc + " is", defaultBandWidth, "kbit")
	defaultBandWidth = defaultBandWidth * (100 - percentage) / 100
	newBandwidth = strconv.Itoa(defaultBandWidth)
	cmd := "sudo tc qdisc add dev eth0 root tbf rate "
	cmd += newBandwidth + "kbit burst 50000kb latency 500ms"
	fmt.Println("Command:", cmd)
	execCmd(cmd)
}
