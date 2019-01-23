package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	CreateOrLoadBlockchaindb()
	getLocalPeer()

	cli := CLI{}
	cli.Run()
}

func getLocalPeer() {
	ips, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var localPeer string

	for _, ipAddr := range ips {
		if ipnet, ok := ipAddr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localPeer = ipnet.IP.String()
				break
			}
		}
	}
	LocalPeer = localPeer + ":3000"
}
