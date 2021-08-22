package main

import (
	"fmt"
	"github.com/go-ping/ping"
	"os"
	"time"
)

func main() {
	Ping("8.8.8.8")
}
func Ping(host string) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

	//// listen for ctrl-C signal
	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt)
	//go func() {
	//	for range c {
	//		pinger.Stop()
	//	}
	//}()

	lastSequence :=-1
	pinger.OnRecv = func(pkt *ping.Packet) {
		if pkt.Seq!=lastSequence+1 {
			missed:=pkt.Seq-lastSequence-1
			fmt.Printf("Missed %d packets!\n", missed)
			f, err := os.OpenFile("pinger.csv",os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				fmt.Printf("openfile error %v\n", err)
				return
			}
			defer f.Close()
			fmt.Fprintf(f, "%s;missed;%d\n", time.Now().Format("2006-01-02 15:04:05"), missed)
		}
		lastSequence=pkt.Seq
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}
	//pinger.OnSend = func(pkt *ping.Packet) {
	//	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
	//		pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	//}
	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
		pinger.OnRecv(pkt)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d packets transmitted, %d packets received, %d duplicates, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketsRecvDuplicates, stats.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	//pinger.Count = 10
	pinger.Size = 32
	//pinger.Interval = 1
	//pinger.Timeout = 2*time.Second
	pinger.SetPrivileged(true)

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	err = pinger.Run()
	if err != nil {
		fmt.Printf("Failed to ping target host: %s", err)
	}
}
