package main

import (
	"flag"
	"os"

	dhcpv4d "github.com/hicompute/histack/pkg/daemon/dhcp"
	"k8s.io/klog/v2"
)

func main() {
	var ifaceName string
	var serverAddress string
	var ipFamily string
	var macPrefix string

	flag.StringVar(&ifaceName, "iface", "br-ext", "The interface to listen for dhcp packets.")
	flag.StringVar(&serverAddress, "server-address", "0.0.0.0", "The server address.")
	flag.StringVar(&ipFamily, "ip-family", "v4", "v4/v6")
	flag.StringVar(&macPrefix, "mac-prefix", "02", "only effects on mac addresses with specific prefix. default 02 ")

	os.Setenv("MAC_PREFIX", macPrefix)
	if ipFamily == "v4" {
		os.Setenv("HISTACK_DHCP4_SERVER_ADDRESS", serverAddress)
		if err := dhcpv4d.Start(ifaceName); err != nil {
			klog.Fatalf("Error on starting dhcp daemon: %v", err)
		}
	}
}
