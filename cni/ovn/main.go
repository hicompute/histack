package main

import (
	ovncnid "github.com/hicompute/histack/pkg/daemon/ovn-cni-server"
)

func main() {
	if err := ovncnid.Start("/var/run/histack-ovn-cni.sock"); err != nil {
		panic(err)
	}
}
