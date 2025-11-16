package main

import (
	"flag"

	ovncnid "github.com/hicompute/histack/pkg/daemon/ovn-cni-server"
	"k8s.io/klog/v2"
)

func main() {
	var cniSocketFile string

	flag.StringVar(&cniSocketFile, "cni-socket", "/var/run/histack-ovn-cni.sock", "The unix socket file cni daemon should create.")
	if err := ovncnid.Start("/var/run/histack-ovn-cni.sock"); err != nil {
		klog.Fatalf("Error on starting ovn cni daemon: %v", err)
	}
}
