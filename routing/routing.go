package routing

import (
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

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
