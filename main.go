package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	ipAddrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, ipAddr := range ipAddrs {
		if ipnet, ok := ipAddr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localNodeAddress = ipnet.IP.String()
				break
			}
		}
	}

	localNodeAddress = fmt.Sprintf(localNodeAddress + ":3000")
	CreateOrLoadBlockChaindb()

	cli := CLI{}
	cli.Run()
}
