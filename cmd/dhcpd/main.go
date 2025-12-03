package main

import (
	dhcpv4d "github.com/hicompute/histack/pkg/daemon/dhcp"
	"k8s.io/klog/v2"
)

func main() {
	if err := dhcpv4d.Start(); err != nil {
		klog.Fatalf("Error on starting dhcp daemon: %v", err)
	}
}
